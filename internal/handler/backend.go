package handler

import (
    "net/http"
    "gopanel/internal/util"
	"github.com/gorilla/sessions"
    "encoding/json"
	"fmt"
	"io"
	"log"
)

var (
	randomString, _ = util.GenerateSecretKey(16)
    key   = []byte(randomString)
    store = sessions.NewCookieStore(key)
	qrCode, Secret, err = util.Generate2FAQRCode()
)



type PinRequest struct {
    PIN string `json:"pin"`
}



var otpFile = "/etc/gopanel"

func BackendHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is the backend response from port 1337!")
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]bool{
		"nginx": util.CheckIfInstalled("nginx"),
		"php8.1-fpm":   util.CheckIfInstalled("php8.1-fpm"),
		"mariadb-server": util.CheckIfInstalled("mariadb-server"),
		"cron":  util.CheckIfInstalled("cron"),
	}	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func StackInstallationHandler(w http.ResponseWriter, r *http.Request) {
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
    util.CheckAndInstallSoftware(requestData.Packages)

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Installation process initiated for packages: %v", requestData.Packages)
}

func Generate2FAHandler(w http.ResponseWriter, r *http.Request) {
	if util.FileExists(otpFile) {
        log.Printf("OTP file exists")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("OTP already exists"))
    } else {
		
		if err != nil {
			http.Error(w, "Error generating QR code", http.StatusInternalServerError)
			return
		}
	
		// Send QR code image as response
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "inline; filename=\"qrcode.png\"")
		w.Write(qrCode)
	
		// Optionally, log the OTP secret (for testing purposes)
		log.Println("OTP Secret:", Secret)
		// file, err := os.OpenFile("/etc/gopanel", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		// defer file.Close()
	
		// // Write the content to the file
		// _, err = file.WriteString(secret)

    }


}

func ValidateOTPHandler(w http.ResponseWriter, r *http.Request) {
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
	valid, err := util.ValidateOTP(pinReq.PIN, Secret)
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
		util.Write2FA(Secret)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid OTP"))
	}
	log.Println(r.Method, r.URL, r.RemoteAddr)
}
