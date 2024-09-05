package server

import (
    "net/http"
	"log"
    "gopanel/internal/handler"
)

func WithCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allowing specific origins, methods, and headers
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8888") // Change "*" to a specific origin if needed
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")


		log.Println(r.Method, r.URL, r.RemoteAddr)
        // Handle OPTIONS method for preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        session, _ := handler.Store.Get(r, "session")

        // Call the next handler if it's not an OPTIONS request
        if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
