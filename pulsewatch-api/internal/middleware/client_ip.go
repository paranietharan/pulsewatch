package middleware

import (
	"net"
	"net/http"
	"strings"

	"pulsewatch-api/internal/config"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// Direct alias exports to standard Chi v5.3+ client IP middlewares
var (
	// ClientIPFromRemoteAddr reads IP directly from TCP RemoteAddr without proxy headers.
	ClientIPFromRemoteAddr = chimw.ClientIPFromRemoteAddr
	// ClientIPFromHeader stores client IP from a single-IP header (e.g. CF-Connecting-IP).
	ClientIPFromHeader = chimw.ClientIPFromHeader
	// ClientIPFromXFF stores client IP from X-Forwarded-For walking past trusted CIDRs.
	ClientIPFromXFF = chimw.ClientIPFromXFF
	// ClientIPFromXFFTrustedProxies stores client IP walking back n trusted proxy hops in XFF.
	ClientIPFromXFFTrustedProxies = chimw.ClientIPFromXFFTrustedProxies
)

// SetupClientIP returns the recommended Chi v5.3+ ClientIP middleware based on deployment configuration:
//   - Direct internet access, no reverse proxy: chimw.ClientIPFromRemoteAddr
//   - Behind Cloudflare:                       chimw.ClientIPFromHeader("CF-Connecting-IP")
//   - Custom trusted header:                   chimw.ClientIPFromHeader(header)
//   - Behind a proxy with known IP ranges:     chimw.ClientIPFromXFF(cidrs...)
//   - Fixed number of trusted proxy hops:      chimw.ClientIPFromXFFTrustedProxies(n)
func SetupClientIP(cfg *config.Config) func(http.Handler) http.Handler {
	if cfg == nil {
		return chimw.ClientIPFromRemoteAddr
	}

	strategy := strings.ToLower(strings.TrimSpace(cfg.ClientIPStrategy))
	switch strategy {
	case "cloudflare":
		header := cfg.ClientIPHeader
		if header == "" {
			header = "CF-Connecting-IP"
		}
		return chimw.ClientIPFromHeader(header)

	case "header":
		header := cfg.ClientIPHeader
		if header == "" {
			header = "X-Real-IP"
		}
		return chimw.ClientIPFromHeader(header)

	case "xff", "xff_cidrs":
		if len(cfg.ClientIPTrustedCIDRs) > 0 {
			return chimw.ClientIPFromXFF(cfg.ClientIPTrustedCIDRs...)
		}
		return chimw.ClientIPFromXFF()

	case "proxy_hops", "trusted_proxies":
		hops := cfg.ClientIPTrustedProxies
		if hops <= 0 {
			hops = 1
		}
		return chimw.ClientIPFromXFFTrustedProxies(hops)

	case "remote_addr":
		fallthrough
	default:
		return chimw.ClientIPFromRemoteAddr
	}
}

// GetClientIP retrieves the resolved client IP address stored in request context by Chi's ClientIP middlewares.
// If context does not have a client IP (e.g. middleware unmounted or test mock), falls back to parsing r.RemoteAddr.
func GetClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	if ip := chimw.GetClientIP(r.Context()); ip != "" {
		return ip
	}

	// Fallback: extract host from r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
