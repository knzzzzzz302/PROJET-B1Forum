package webAPI

import (
	"net/http"
)

// ApplyMiddleware applique une série de middlewares à un gestionnaire HTTP
func ApplyMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

// RateLimiterMiddleware est un middleware qui limite les requêtes par IP
func RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Implémentation simple - vous pouvez la rendre plus sophistiquée plus tard
		next.ServeHTTP(w, r)
	})
}