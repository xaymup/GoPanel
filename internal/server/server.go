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
    backendMux.HandleFunc("/api/load", util.LoadHandler)
	backendMux.HandleFunc("/api/disk-usage", util.DiskUsageHandler)
    backendMux.HandleFunc("/api/resource-usage", util.ResourceUtilHandler)
    backendMux.HandleFunc("/api/list-sites", util.ListSites)
    backendMux.HandleFunc("/api/write-siteconf", util.WriteSiteConf)
    backendMux.HandleFunc("/api/list-files", util.ListFiles)
    backendMux.HandleFunc("/api/upload-file", util.UploadFile)
    backendMux.HandleFunc("/api/rename-file", util.RenameFile)
    backendMux.HandleFunc("/api/copy-file", util.CopyFile)
    backendMux.HandleFunc("/api/download-file", util.DownloadFile)
    backendMux.HandleFunc("/api/delete-file", util.DeleteFile)
    backendMux.HandleFunc("/api/compress-file", util.CompressFile)
    backendMux.HandleFunc("/api/extract-file", util.ExtractFile)
    backendMux.HandleFunc("/api/get-file", util.GetFile)
    backendMux.HandleFunc("/api/update-file", util.UpdateFile)



    frontendMux := http.NewServeMux()
    frontendMux.HandleFunc("/", handler.FrontendHandler)
    frontendMux.HandleFunc("/validate-otp", handler.ValidateOTPHandler)
    frontendMux.HandleFunc("/api/status", handler.StatusHandler)
    frontendMux.HandleFunc("/api/install-stack", handler.StackInstallationHandler)
    frontendMux.HandleFunc("/api/generate-2fa.png", handler.Generate2FAHandler)

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
