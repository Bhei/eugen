package eugen

import "net/http"

func AddSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Prodection", "1; mode=block")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
}

func AddSaneHTMLHeaders(w http.ResponseWriter) {
	//Disable Caching
	w.Header().Set("Cache-control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Keep-Alive", "timeout=5, max=100")
}
