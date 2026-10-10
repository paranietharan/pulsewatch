package ssrf

import (
	"net"
	"net/http"
	"net/url"
	"testing"
)

func TestIsPrivateOrBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.10", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true}, // Cloud metadata
		{"0.0.0.0", true},
		{"8.8.8.8", false},       // Google DNS
		{"1.1.1.1", false},       // Cloudflare DNS
		{"93.184.216.34", false}, // example.com
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			parsed := net.ParseIP(tt.ip)
			if parsed == nil {
				t.Fatalf("failed to parse IP %s", tt.ip)
			}
			result := IsPrivateOrBlockedIP(parsed)
			if result != tt.blocked {
				t.Errorf("expected blocked=%t for %s, got %t", tt.blocked, tt.ip, result)
			}
		})
	}
}

func TestValidateTargetURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid public HTTPS", "https://api.github.com/status", false},
		{"Valid public HTTP", "http://example.com/test", false},
		{"Blocked file scheme", "file:///etc/passwd", true},
		{"Blocked gopher scheme", "gopher://127.0.0.1:70", true},
		{"Blocked embedded user", "http://admin:secret@example.com", true},
		{"Blocked localhost IP", "http://127.0.0.1:8080/metrics", true},
		{"Blocked private IP", "http://192.168.1.100/admin", true},
		{"Blocked cloud metadata", "http://169.254.169.254/latest/meta-data/", true},
		{"Empty URL", "", true},
		{"Invalid URL format", ":::bad-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateTargetURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTargetURL(%q) err = %v, wantErr = %t", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestSafeCheckRedirect(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://169.254.169.254/secret", nil)
	via := []*http.Request{{URL: &url.URL{Scheme: "https", Host: "example.com"}}}

	err := SafeCheckRedirect(req, via)
	if err == nil {
		t.Errorf("expected redirect to metadata IP to be blocked, but succeeded")
	}
}
