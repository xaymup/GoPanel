package util

import (
    "net/http"
	"fmt"
	"io/ioutil"
)

func GetServerIP() (string, error) {
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