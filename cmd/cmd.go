package cmd

import (
    "log"
    "flag"
)

var development = flag.Bool("development", false, "Run in development mode")


func GetMode() (bool) {
	// Parse flags




    flag.Parse()
	// Use the flags

    if (*development){
        log.Println("Running in development mode")
    } else {
        log.Println("Not running in development mode")
    }

    return *development
}