package main

import (
    "gopanel/internal/server"
    "log"
)


func main() {

    // Parse the flags

    log.Println("Starting GoPanel 0.1")

    server.Start()
}

