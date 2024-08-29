package main

import (
    "embed"
    "fmt"
    "log"
    "net/http"
)

//go:embed static/index.html
var content embed.FS

func main() {
    // Backend handler
    backendHandler := func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "This is the backend response from port 1337!")
    }

    // Create and start the backend server
    go func() {
        backendMux := http.NewServeMux()
        backendMux.HandleFunc("/", backendHandler)
        port := ":1337"
        log.Printf("Starting backend server on port %s...", port)
        if err := http.ListenAndServe(port, backendMux); err != nil {
            log.Fatalf("Failed to start backend server: %v", err)
        }
    }()

    // Frontend handler using embedded content
    frontendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        data, err := content.ReadFile("static/index.html")
        if err != nil {
            http.Error(w, "File not found", http.StatusNotFound)
            return
        }
        w.Header().Set("Content-Type", "text/html")
        w.Write(data)
    })

    // Create and start the frontend server
    go func() {
        http.Handle("/", frontendHandler)
        port := ":8888"
        log.Printf("Starting frontend server on port %s...", port)
        if err := http.ListenAndServe(port, nil); err != nil {
            log.Fatalf("Failed to start frontend server: %v", err)
        }
    }()

    // Block the main goroutine to keep the servers running
    select {}
}
