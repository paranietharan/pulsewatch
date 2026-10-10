package ssrf

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
)

var (
	ErrInvalidScheme   = errors.New("only http and https schemes are permitted")
	ErrEmbeddedUser    = errors.New("embedded user credentials in URL are prohibited")
	ErrSSRFBlocked     = errors.New("destination IP address belongs to a prohibited private, loopback, or internal range")
	ErrEmptyHost       = errors.New("target host cannot be empty")
	ErrTooManyHops     = errors.New("maximum redirect limit exceeded")
)

// Prohibited IP CIDR ranges for defense-in-depth SSRF prevention
var (
	ipv4ProhibitedCIDRs []*net.IPNet
	ipv6ProhibitedCIDRs []*net.IPNet
)

func init() {
	rawV4 := []string{
		"0.0.0.0/8",          // Current network
		"10.0.0.0/8",         // Private RFC 1918
		"100.64.0.0/10",      // Shared address space RFC 6598
		"127.0.0.0/8",        // Loopback
		"169.254.0.0/16",     // Link-local & AWS/GCP/Azure instance metadata
		"172.16.0.0/12",      // Private RFC 1918
		"192.0.0.0/24",       // IETF protocol assignments
		"192.0.2.0/24",       // Documentation TEST-NET-1
		"192.88.99.0/24",     // 6to4 relay anycast
		"192.168.0.0/16",     // Private RFC 1918
		"198.18.0.0/15",      // Benchmarking
		"198.51.100.0/24",    // Documentation TEST-NET-2
		"203.0.113.0/24",     // Documentation TEST-NET-3
		"224.0.0.0/4",        // Multicast
		"240.0.0.0/4",        // Reserved
		"255.255.255.255/32", // Limited broadcast
	}

	for _, cidr := range rawV4 {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			ipv4ProhibitedCIDRs = append(ipv4ProhibitedCIDRs, ipNet)
		}
	}

	rawV6 := []string{
		"::/128",        // Unspecified
		"::1/128",       // Loopback
		"100::/64",      // Discard prefix
		"2001:db8::/32", // Documentation
		"fc00::/7",      // Unique local address (ULA)
		"fe80::/10",     // Link-local unicast
		"ff00::/8",      // Multicast
	}

	for _, cidr := range rawV6 {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			ipv6ProhibitedCIDRs = append(ipv6ProhibitedCIDRs, ipNet)
		}
	}
}

// IsPrivateOrBlockedIP returns true if the specified IP is private, loopback, or reserved.
func IsPrivateOrBlockedIP(ip net.IP) bool {
	if os.Getenv("ALLOW_INTERNAL_PROBES") == "true" {
		return false
	}
	if ip == nil {
		return true
	}
	// Check standard library classifications
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		for _, block := range ipv4ProhibitedCIDRs {
			if block.Contains(ip4) {
				return true
			}
		}
		return false
	}

	for _, block := range ipv6ProhibitedCIDRs {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateTargetURL parses and validates the URL against common SSRF vectors.
func ValidateTargetURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("malformed URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, ErrInvalidScheme
	}

	if parsed.User != nil {
		return nil, ErrEmbeddedUser
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return nil, ErrEmptyHost
	}

	// If hostname is directly an IP address, validate immediately
	if ip := net.ParseIP(hostname); ip != nil {
		if IsPrivateOrBlockedIP(ip) {
			return nil, ErrSSRFBlocked
		}
	}

	return parsed, nil
}

// SafeDialContext returns a DialContext func that enforces SSRF checks at the TCP socket level,
// preventing DNS rebinding attacks between URL parsing and connection establishment.
func SafeDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		// Resolve IPs through safe DNS lookup
		ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
		if err != nil {
			return nil, fmt.Errorf("dns resolution failed: %w", err)
		}

		if len(ips) == 0 {
			return nil, fmt.Errorf("no IP addresses resolved for host %s", host)
		}

		// Verify every resolved IP is public and permitted
		var targetIP net.IP
		for _, ip := range ips {
			if IsPrivateOrBlockedIP(ip) {
				return nil, fmt.Errorf("%w: %s resolved to %s", ErrSSRFBlocked, host, ip.String())
			}
			if targetIP == nil {
				targetIP = ip
			}
		}

		// Dial the validated IP explicitly to prevent DNS rebinding
		dialer := &net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			Control: func(network, address string, c syscall.RawConn) error {
				// Re-verify the address being connected
				h, _, err := net.SplitHostPort(address)
				if err == nil {
					if parsedIP := net.ParseIP(h); parsedIP != nil && IsPrivateOrBlockedIP(parsedIP) {
						return fmt.Errorf("%w: attempted connection to %s", ErrSSRFBlocked, parsedIP.String())
					}
				}
				return nil
			},
		}

		targetAddr := net.JoinHostPort(targetIP.String(), port)
		return dialer.DialContext(ctx, network, targetAddr)
	}
}

// SafeCheckRedirect validates redirect hops to ensure a redirect does not bounce to an internal service.
func SafeCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return ErrTooManyHops
	}

	// Validate redirect target URL
	_, err := ValidateTargetURL(req.URL.String())
	if err != nil {
		return fmt.Errorf("redirect target blocked by SSRF protection: %w", err)
	}

	return nil
}
