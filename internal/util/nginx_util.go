package util

import (
	"encoding/json"
	"net/http"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"regexp"
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
		// if strings.HasPrefix(line, "site_name=") {
		// 	response.SiteName = strings.TrimPrefix(line, "site_name=")
		// } else if strings.HasPrefix(line, "domains=") {
		// 	response.Domains = strings.Split(strings.TrimPrefix(line, "domains="), ",")
		// } else if strings.HasPrefix(line, "path=") {
		// 	response.Path = strings.TrimPrefix(line, "path=")
		// } else if strings.HasPrefix(line, "php_version=") {
		// 	response.PHPVersion = strings.TrimPrefix(line, "php_version=")
		// }
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