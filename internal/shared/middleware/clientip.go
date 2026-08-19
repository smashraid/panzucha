package middleware

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP returns middleware that extracts the real client IP
// from X-Forwarded-For when the request comes from a trusted proxy.
func ClientIP(trustedProxies []*net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			remoteIPStr, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				remoteIPStr = r.RemoteAddr
			}
			remoteIP := net.ParseIP(remoteIPStr)

			if !isIPTrusted(remoteIP, trustedProxies) {
				next.ServeHTTP(w, r)
				return
			}

			xff := r.Header.Get("X-Forwarded-For")
			if xff == "" {
				next.ServeHTTP(w, r)
				return
			}

			ips := strings.Split(xff, ",")
			for i := len(ips) - 1; i >= 0; i-- {
				ipStr := strings.TrimSpace(ips[i])
				clientIP := net.ParseIP(ipStr)
				if clientIP == nil {
					continue
				}
				if !isIPTrusted(clientIP, trustedProxies) {
					r.Header.Set("X-Real-IP", ipStr)
					break
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isIPTrusted(ip net.IP, trustedProxies []*net.IPNet) bool {
	if ip.IsLoopback() {
		return true
	}
	for _, cidr := range trustedProxies {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
