package main

import (
	"encoding/json"
    "embed"
    "fmt"
    "log"
    "net/http"
	"os/exec"
	"io"
)

//go:embed static/index.html
var content embed.FS

func withCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allowing specific origins, methods, and headers
        w.Header().Set("Access-Control-Allow-Origin", "*") // Change "*" to a specific origin if needed
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle OPTIONS method for preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Call the next handler if it's not an OPTIONS request
        next.ServeHTTP(w, r)
    })
}

func detectPackageManager() (string, error) {
	// List of common package managers
	packageManagers := []string{
		"apt-get", "apt", "yum", "dnf", "zypper", "pacman", "brew", "apk", "pkg",
	}

	for _, pm := range packageManagers {
		cmd := exec.Command("which", pm)
		err := cmd.Run()
		if err == nil {
			return pm, nil
		}
	}

	return "", fmt.Errorf("no supported package manager found")
}

func pkgManager(action, software string) error {
	pm, err := detectPackageManager()
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	var queryCmd string
	
	switch pm {
	case "apt-get", "apt":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "remove", "-y", software)
		} else if action == "check" {
			queryCmd = fmt.Sprintf("dpkg -l | grep ^ii | awk '{print $2}' | grep ^%s$", software)
			cmd = exec.Command("sh", "-c", queryCmd)
		}
	case "yum":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "remove", "-y", software)
		} else if action == "check" {
			cmd = exec.Command("rpm", "-q", software)
		}
	case "dnf":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "remove", "-y", software)
		} else if action == "check" {
			cmd = exec.Command("dnf", "list", "installed", software)
		}
	case "zypper":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "install", "-y", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "remove", "-y", software)
		} else if action == "check" {
			cmd = exec.Command("zypper", "se", "--installed-only", software)
		}
	case "pacman":
		if action == "install" {
			cmd = exec.Command("sudo", pm, "-S", "--noconfirm", software)
		} else if action == "remove" {
			cmd = exec.Command("sudo", pm, "-R", "--noconfirm", software)
		} else if action == "check" {
			cmd = exec.Command("pacman", "-Q", software)
		}
	case "brew":
		if action == "install" {
			cmd = exec.Command("brew", "install", software)
		} else if action == "remove" {
			cmd = exec.Command("brew", "uninstall", software)
		} else if action == "check" {
			cmd = exec.Command("brew", "list", software)
		}
	case "apk":
		if action == "install" {
			cmd = exec.Command("apk", "add", software)
		} else if action == "remove" {
			cmd = exec.Command("apk", "del", software)
		} else if action == "check" {
			cmd = exec.Command("apk", "info", software)
		}
	case "pkg":
		if action == "install" {
			cmd = exec.Command("pkg", "install", "-y", software)
			log.Printf("installing '%s'", software)
		} else if action == "remove" {
			cmd = exec.Command("pkg", "delete", "-y", software)
		} else if action == "check" {
			cmd = exec.Command("pkg", "info", software)
		}
	default:
		return fmt.Errorf("unsupported package manager: %s", pm)
	}

	if cmd == nil {
		return fmt.Errorf("no command found for package manager: %s", pm)
	}

	if action == "check" {
		_, err := cmd.CombinedOutput()
		if err != nil {
			return err
		} else {
			return nil
		}
	}

	return cmd.Run() // For install and remove actions
}

func checkIfInstalled(serviceName string) bool {
	// Check if the software is installed.
	err := pkgManager("check", serviceName)
	if err == nil {
		return true
	} else {
		return false
	}
}

func backendHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is the backend response from port 1337!")

}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]bool{
		"nginx": checkIfInstalled("nginx"),
		"php8.1-fpm":   checkIfInstalled("php8.1-fpm"),
		"mariadb-server": checkIfInstalled("mariadb-server"),
		"cron":  checkIfInstalled("cron"),
	}	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func checkAndInstallSoftware(packages []string) {
    for _, pkg := range packages {
        if !checkIfInstalled(pkg) {
			err := pkgManager("install", pkg)
			fmt.Printf("installing: %s \n", pkg)
			if err != nil {
				fmt.Println("Error:", err)
			}
        } else {
            fmt.Printf("%s is already installed.\n", pkg)
        }
    }
}

func stackInstallationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var requestData struct {
        Packages []string `json:"packages"`
    }

    // Parse the request body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusInternalServerError)
        return
    }

    err = json.Unmarshal(body, &requestData)
    if err != nil {
        http.Error(w, "Error parsing JSON", http.StatusBadRequest)
        return
    }
    // Check and install the requested software
    checkAndInstallSoftware(requestData.Packages)

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Installation process initiated for packages: %v", requestData.Packages)
}



func main() {
    // Backend handler


    // Create and start the backend server
    go func() {
        backendMux := http.NewServeMux()
        backendMux.HandleFunc("/api", backendHandler)
		backendMux.HandleFunc("/api/status", statusHandler)
		backendMux.HandleFunc("/api/install-stack", stackInstallationHandler)
        port := ":1337"
        log.Printf("Starting backend server on port %s...", port)
        if err := http.ListenAndServe(port, withCORS(backendMux)); err != nil {
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
