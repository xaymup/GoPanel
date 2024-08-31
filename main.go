package main

import (
	"encoding/json"
    "embed"
    "fmt"
    "log"
    "net/http"
	"os/exec"
	"io"
    "path/filepath"
	"github.com/pquerna/otp/totp"
    "github.com/skip2/go-qrcode"
	"io/ioutil"
	"os"
	"github.com/gorilla/sessions"
    "crypto/rand"
    "encoding/base64"
	"github.com/shirou/gopsutil/load"
	"runtime"
	"html/template"
)

// Embedding filesystem
//go:embed static/*
var content embed.FS

// Initializing templates
var templates = template.Must(template.ParseGlob("static/*.html"))

type LoadAvg struct {
	Cores  int 	   `json:"cores"`
    Load1  float64 `json:"load1"`
    Load5  float64 `json:"load5"`
    Load15 float64 `json:"load15"`
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
    avg, err := load.Avg()
    if err != nil {
        http.Error(w, "Could not retrieve load average", http.StatusInternalServerError)
        return
    }

	numCPU := runtime.NumCPU()

    loadAvg := LoadAvg{
		Cores:  numCPU,
        Load1:  avg.Load1,
        Load5:  avg.Load5,
        Load15: avg.Load15,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(loadAvg)

}


var otpFile = "/etc/gopanel"

func generateSecretKey(length int) (string, error) {
    // Create a byte slice with the desired length
    bytes := make([]byte, length)

    // Fill the byte slice with random data
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }

    // Encode the byte slice to a base64 string
    // Using base64 encoding to avoid issues with non-printable characters
    randomString := base64.URLEncoding.EncodeToString(bytes)

    // Truncate the string to the desired length
    if length < len(randomString) {
        randomString = randomString[:length]
    }

    return randomString, nil
}

var (
	randomString, _ = generateSecretKey(16)
    key   = []byte(randomString)
    store = sessions.NewCookieStore(key)
)

func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    if err == nil {
        return true
    }
    if os.IsNotExist(err) {
        return false
    }
    return false
}

type PinRequest struct {
    PIN string `json:"pin"`
}

func readSecretFromFile() (string, error) {
	filePath := "/etc/gopanel"
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", fmt.Errorf("error reading file: %w", err)
    }
    return string(data), nil
}

func validateOTP(otp string) (bool, error) {

    // Read the secret from the file
    secret, err := readSecretFromFile()
    if err != nil {
        return false, err
    }

    // Validate the OTP using the secret
    valid := totp.Validate(otp, secret)
    return valid, nil
}

func getServerIP() (string, error) {
    // URL of the service that provides the public IP address
    url := "https://ifconfig.me"

    // Perform the HTTP GET request
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    // Check if the request was successful
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to get IP address, status code: %d", resp.StatusCode)
    }

    // Read the response body
    ipAddress, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(ipAddress), nil
}

func generate2FAQRCode() ([]byte, string, error) {
    // Generate a new OTP key
	ip, _ := getServerIP() 
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "GoPanel",
        AccountName: fmt.Sprintf("admin@%s",ip),
    })
    if err != nil {
        return nil, "", err
    }

    // Generate a URL to encode in the QR code
    url := key.URL()

    // Generate the QR code image
    qrCode, err := qrcode.Encode(url, qrcode.Medium, 256)
    if err != nil {
        return nil, "", err
    }

    return qrCode, key.Secret(), nil
}

func withCORS(next http.Handler) http.Handler {
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
	allowedPackages := map[string]struct{}{
        "nginx":           {},
        "mariadb-server":  {},
        "php8.1-fpm":      {},
        "cron":            {},
    }
		for _, pkg := range packages {
			if _, ok := allowedPackages[pkg]; ok {
				if !checkIfInstalled(pkg) {
					err := pkgManager("install", pkg)
					log.Printf("installing: %s \n", pkg)
					if err != nil {
						log.Printf("Error: %s", err)
					}
					} else {
					log.Printf("%s is already installed.\n", pkg)
				}
			}  else {
				log.Printf("%s is not in the allowed list.\n", pkg)
			}
		}
}

func stackInstallationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }



    // Parse the request body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusInternalServerError)
        return
    }

	var requestData struct {
        Packages []string `json:"packages"`
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

func checkIfStackReady () (bool) {
	if checkIfInstalled("nginx") && checkIfInstalled("mariadb-server") && checkIfInstalled("php8.1-fpm") && checkIfInstalled("cron") {
		return true
	} else {
		return false
	}
}

func generate2FAHandler(w http.ResponseWriter, r *http.Request) {
	

	if fileExists(otpFile) {
        log.Printf("OTP file exists")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("OTP already exists"))
    } else {
		qrCode, secret, err := generate2FAQRCode()
		if err != nil {
			http.Error(w, "Error generating QR code", http.StatusInternalServerError)
			return
		}
	
		// Send QR code image as response
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "inline; filename=\"qrcode.png\"")
		w.Write(qrCode)
	
		// Optionally, log the OTP secret (for testing purposes)
		log.Println("OTP Secret:", secret)
		file, err := os.OpenFile("/etc/gopanel", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		defer file.Close()
	
		// Write the content to the file
		_, err = file.WriteString(secret)
    }

}



func main() {
    // Backend handler


    // Create and start the backend server
    go func() {
        backendMux := http.NewServeMux()
        backendMux.HandleFunc("/api", backendHandler)
		backendMux.HandleFunc("/api/status", statusHandler)
		backendMux.HandleFunc("/api/install-stack", stackInstallationHandler)
		backendMux.HandleFunc("/api/generate-2fa.png", generate2FAHandler)
		backendMux.HandleFunc("/api/load", loadHandler)
        port := ":1337"
        log.Printf("Starting backend server on port %s...", port)
        if err := http.ListenAndServe(port, withCORS(backendMux)); err != nil {
            log.Fatalf("Failed to start backend server: %v", err)
        }
    }()

    // Frontend handler using embedded content
    frontendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileToServe := ""
		if checkIfStackReady() {
			session, _ := store.Get(r, "session")
			filePath := filepath.Join("static", r.URL.Path[1:])
			
			// Check if the user is authenticated
			if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
				if !fileExists(otpFile) {
					fileToServe = "static/account.html"
				} else {
					fileToServe = "static/login.html"
				}
			} else {
				fileToServe = filePath
			}
		} else {
			fileToServe = "static/install.html"
		}
		input := map[string]interface{}{
			"Title": "Home Page",
		}
		err := templates.ExecuteTemplate(w, fileToServe, input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
        // data, err := content.ReadFile(fileToServe)
        if err != nil {
            http.Error(w, "File not found", http.StatusNotFound)
            return
        }
        // w.Header().Set("Content-Type", "text/html")
        // w.Write(data)
		log.Println(r.Method, r.URL, r.RemoteAddr)
    })

	validateOTPHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
	
		var pinReq PinRequest
	
		// Decode the incoming JSON request
		err := json.NewDecoder(r.Body).Decode(&pinReq)
		if err != nil {
			http.Error(w, "Error parsing JSON request", http.StatusBadRequest)
			return
		}
	
		// Validate the OTP/PIN
		valid, err := validateOTP(pinReq.PIN)
		if err != nil {
			http.Error(w, "Error validating OTP", http.StatusInternalServerError)
			return
		}
	
		// Respond based on the validity of the OTP
		if valid {
			session, _ := store.Get(r, "session")
			session.Values["authenticated"] = true
			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   3600, // 1 hour
				HttpOnly: false, // Prevents JavaScript access
				Secure:   false, // Set to true if using HTTPS
				SameSite: http.SameSiteLaxMode, // Adjust as needed
			}
	
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OTP is valid!"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid OTP"))
		}
		log.Println(r.Method, r.URL, r.RemoteAddr)
	})

    // Create and start the frontend server
    go func() {
        http.Handle("/", frontendHandler)
		http.Handle("/validate-otp", validateOTPHandler)
        port := ":8888"
        log.Printf("Starting frontend server on port %s...", port)
        if err := http.ListenAndServe(port, nil); err != nil {
            log.Fatalf("Failed to start frontend server: %v", err)
        }
    }()

    // Block the main goroutine to keep the servers running
    select {}
}
