package server

import (
    "log"
    "net/http"
    "gopanel/internal/handler"
    "gopanel/internal/util"
)

func Start() {
    backendMux := http.NewServeMux()
    backendMux.HandleFunc("/api", handler.BackendHandler)
    backendMux.HandleFunc("/api/status", handler.StatusHandler)
    backendMux.HandleFunc("/api/install-stack", handler.StackInstallationHandler)
    backendMux.HandleFunc("/api/generate-2fa.png", handler.Generate2FAHandler)
    backendMux.HandleFunc("/api/load", util.LoadHandler)

    frontendMux := http.NewServeMux()
    frontendMux.HandleFunc("/", handler.FrontendHandler)
    frontendMux.HandleFunc("/validate-otp", handler.ValidateOTPHandler)

    go func() {
        port := ":1337"
        log.Printf("Starting backend server on port %s...", port)
        if err := http.ListenAndServe(port, WithCORS(backendMux)); err != nil {
            log.Fatalf("Failed to start backend server: %v", err)
        }
    }()

    go func() {
        port := ":8888"
        log.Printf("Starting frontend server on port %s...", port)
        if err := http.ListenAndServe(port, frontendMux); err != nil {
            log.Fatalf("Failed to start frontend server: %v", err)
        }
    }()

	
	select {}

}
