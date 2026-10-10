package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pulsewatch-api/internal/config"
)

func TestClientIPFromRemoteAddr(t *testing.T) {
	handler := ClientIPFromRemoteAddr(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		if ip != "203.0.113.195" {
			t.Errorf("expected 203.0.113.195, got %s", ip)
		}
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "203.0.113.195:54321"
	// Spoofed X-Forwarded-For should be ignored in direct connection mode
	req.Header.Set("X-Forwarded-For", "198.51.100.1")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
}

func TestClientIPFromHeaderCloudflare(t *testing.T) {
	mw := ClientIPFromHeader("CF-Connecting-IP")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		if ip != "198.51.100.5" {
			t.Errorf("expected 198.51.100.5 from CF-Connecting-IP, got %s", ip)
		}
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "172.69.1.1:12345"
	req.Header.Set("CF-Connecting-IP", "198.51.100.5")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
}

func TestClientIPFromXFF(t *testing.T) {
	// Trusted proxy CIDR is 10.0.0.0/8
	mw := ClientIPFromXFF("10.0.0.0/8")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		// Chain: client (203.0.113.50) -> intermediate trusted proxy (10.0.0.1)
		// Walking backwards from right: 10.0.0.1 is in 10.0.0.0/8 (trusted), so client is 203.0.113.50
		if ip != "203.0.113.50" {
			t.Errorf("expected 203.0.113.50 from XFF, got %s", ip)
		}
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:443"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 10.0.0.1")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
}

func TestClientIPFromXFFTrustedProxies(t *testing.T) {
	// 1 trusted hop behind reverse proxy: XFF contains single client IP added by proxy
	mwSingle := ClientIPFromXFFTrustedProxies(1)
	handlerSingle := mwSingle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		if ip != "203.0.113.99" {
			t.Errorf("expected 203.0.113.99, got %s", ip)
		}
	}))

	reqSingle := httptest.NewRequest("GET", "/test", nil)
	reqSingle.RemoteAddr = "127.0.0.1:8080"
	reqSingle.Header.Set("X-Forwarded-For", "203.0.113.99")
	wSingle := httptest.NewRecorder()
	handlerSingle.ServeHTTP(wSingle, reqSingle)

	// 2 trusted hops: Client -> Proxy1 -> Proxy2 -> App
	mwDouble := ClientIPFromXFFTrustedProxies(2)
	handlerDouble := mwDouble(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		if ip != "198.51.100.77" {
			t.Errorf("expected 198.51.100.77, got %s", ip)
		}
	}))

	reqDouble := httptest.NewRequest("GET", "/test", nil)
	reqDouble.RemoteAddr = "127.0.0.1:8080"
	reqDouble.Header.Set("X-Forwarded-For", "198.51.100.77, 10.0.0.2")
	wDouble := httptest.NewRecorder()
	handlerDouble.ServeHTTP(wDouble, reqDouble)
}

func TestSetupClientIPStrategies(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Config
		setupReq   func(r *http.Request)
		expectedIP string
	}{
		{
			name: "Direct remote_addr",
			cfg: &config.Config{
				ClientIPStrategy: "remote_addr",
			},
			setupReq: func(r *http.Request) {
				r.RemoteAddr = "192.0.2.1:80"
			},
			expectedIP: "192.0.2.1",
		},
		{
			name: "Cloudflare strategy",
			cfg: &config.Config{
				ClientIPStrategy: "cloudflare",
				ClientIPHeader:   "CF-Connecting-IP",
			},
			setupReq: func(r *http.Request) {
				r.RemoteAddr = "104.16.1.1:443"
				r.Header.Set("CF-Connecting-IP", "198.51.100.22")
			},
			expectedIP: "198.51.100.22",
		},
		{
			name: "Custom header strategy",
			cfg: &config.Config{
				ClientIPStrategy: "header",
				ClientIPHeader:   "X-Real-IP",
			},
			setupReq: func(r *http.Request) {
				r.RemoteAddr = "127.0.0.1:80"
				r.Header.Set("X-Real-IP", "198.51.100.33")
			},
			expectedIP: "198.51.100.33",
		},
		{
			name: "Proxy hops strategy",
			cfg: &config.Config{
				ClientIPStrategy:       "proxy_hops",
				ClientIPTrustedProxies: 1,
			},
			setupReq: func(r *http.Request) {
				r.RemoteAddr = "127.0.0.1:80"
				r.Header.Set("X-Forwarded-For", "198.51.100.44")
			},
			expectedIP: "198.51.100.44",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := SetupClientIP(tt.cfg)
			var capturedIP string
			h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedIP = GetClientIP(r)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			tt.setupReq(req)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if capturedIP != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, capturedIP)
			}
		})
	}
}

func TestGetClientIPFallback(t *testing.T) {
	// Without middleware executed, GetClientIP parses host from RemoteAddr
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.50:9999"
	ip := GetClientIP(req)
	if ip != "192.168.1.50" {
		t.Errorf("expected 192.168.1.50, got %s", ip)
	}

	// Empty request
	if empty := GetClientIP(nil); empty != "" {
		t.Errorf("expected empty string for nil request, got %s", empty)
	}
}
