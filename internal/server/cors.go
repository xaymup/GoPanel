package server

import (
    "net/http"
	"log"
)

func WithCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allowing specific origins, methods, and headers
        w.Header().Set("Access-Control-Allow-Origin", "*") // Change "*" to a specific origin if needed
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		log.Println(r.Method, r.URL, r.RemoteAddr)
        // Handle OPTIONS method for preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }


        // Call the next handler if it's not an OPTIONS request
        next.ServeHTTP(w, r)
    })
}
