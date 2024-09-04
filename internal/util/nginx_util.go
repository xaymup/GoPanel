package util

import (
	"encoding/json"
	"net/http"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"regexp"
	"log"
	"os"
	"os/exec"
)

var nginxDir = "/etc/nginx/sites-enabled/"

type NginxSite struct {
	SiteName   string   `json:"site_name"`
	Domains    []string `json:"domains"`
	Path       string   `json:"path"`
	PHPVersion string   `json:"php_version"`
}

func ParseSiteConfig(filePath string) (NginxSite, error) {
	var response NginxSite

	log.Println("Parsing NGINX configuration")


	// Read the file content
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		// http.Error(http.StatusInternalServerError, err)
		return response, err
	}

	// Simple parsing logic (customize this based on your configuration format)
	lines := strings.Split(string(data), "\n")
	response.SiteName = strings.ReplaceAll(filepath.Base(filePath), "_", " ")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "server_name") {
			line = strings.TrimPrefix(line, "server_name")
			line = strings.TrimSuffix(line, ";")
			line = strings.TrimSpace(line)
			response.Domains = strings.Fields(line)
		}
		if strings.HasPrefix(line, "root") {
			line = strings.TrimPrefix(line, "root")
			line = strings.TrimSuffix(line, ";")
			line = strings.TrimSpace(line)
			response.Path = line
		}
		if strings.HasPrefix(line, "fastcgi_pass") {
			re := regexp.MustCompile(`php([0-9]+\.[0-9]+)-fpm.sock`)

			// Find the first match for the pattern in the line
			matches := re.FindStringSubmatch(line)
		
			// Check if a match was found
			if len(matches) > 1 {
				// Return the captured version (second element in matches)
				response.PHPVersion = matches[1]
			}
		}
	}

	return response, nil
}

// Handler function to list files in /etc/nginx/sites-available/
func ListSites(w http.ResponseWriter, r *http.Request) {

	var responses []NginxSite

	files, err := ioutil.ReadDir(nginxDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(nginxDir, file.Name())
			response, err := ParseSiteConfig(filePath)
			if err != nil {
				fmt.Printf("Error parsing file %s: %v\n", file.Name(), err)
				continue
			}
			responses = append(responses, response)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}


func WriteNginxConf (sites []NginxSite) error {
	RemoveAllFiles(nginxDir)
	for _, site := range sites {
        // Replace spaces in SiteName with underscores
        siteName := strings.ReplaceAll(site.SiteName, " ", "_")
        
        // Join domains with a comma
        domains := strings.Join(site.Domains, " ")
        
        // Define the content of the configuration file
        configContent := fmt.Sprintf(`
server {
    listen 80;
    server_name %s;

    root %s;
    index index.php index.html index.htm;

    # Log files
    access_log /var/log/nginx/%s_access.log;
    error_log /var/log/nginx/%s_error.log;

    # Handle requests for static files
    location / {
        try_files $uri $uri/ /index.php?$args;
    }

    # Handle PHP files
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php%s-fpm.sock; # Adjust PHP version as needed
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }

    # Disable access to hidden files and directories
    location ~ /\. {
        deny all;
    }

    # Cache control for static files
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|pdf)$ {
        expires 1d;
        log_not_found off;
    }
}
`, domains, site.Path, siteName, siteName, site.PHPVersion)

        // Specify the file path
        filePath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", siteName)
        
        // Create or overwrite the file
        file, err := os.Create(filePath)
        if err != nil {
            return fmt.Errorf("could not create file for site %s: %v", site.SiteName, err)
        }
        defer file.Close()
        
        // Write the content to the file
        _, err = file.WriteString(configContent)
        if err != nil {
            return fmt.Errorf("could not write to file for site %s: %v", site.SiteName, err)
        }

        fmt.Printf("Configuration file written successfully for site: %s\n", site.SiteName)
    }
    restartNginx()
    return nil

}


func WriteSiteConf (w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

	var sites []NginxSite

    // Parse the request body
	err := json.NewDecoder(r.Body).Decode(&sites)



    if err != nil {
        http.Error(w, "Error reading request body", http.StatusInternalServerError)
        return
    }

    if err != nil {
        http.Error(w, "Error parsing JSON", http.StatusBadRequest)
        return
    }
    // Check and install the requested software
    WriteNginxConf(sites)
}

func restartNginx() error {
	// Define the command to restart Nginx
	cmd := exec.Command("sudo", "systemctl", "restart", "nginx")

	// Set the command output to be displayed in the terminal
	cmd.Stdout = exec.Command("tee", "/dev/stdout").Stdout
	cmd.Stderr = exec.Command("tee", "/dev/stderr").Stderr

	log.Println("Restarting Nginx")

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart Nginx: %v", err)
	}

	return nil
}