package util

import (
    "os"
	"fmt"
	"io/ioutil"
	"log"
)

func FileExists(path string) bool {
    _, err := os.Stat(path)
    return !os.IsNotExist(err)
}

func Write2FA( Secret string) error{
	if FileExists("/etc/gopanel"){
        return nil
	}

	file,_ := os.OpenFile("/etc/gopanel", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer file.Close()

	log.Println("Writing PIN token to /etc/gopanel")


	// Write the content to the file
	file.WriteString(Secret)
	return nil
}

func ReadSecretFromFile() (string, error) {
	filePath := "/etc/gopanel"
	log.Println("Reading PIN token file /etc/gopanel")

    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", fmt.Errorf("error reading file: %w", err)
    }
    return string(data), nil
}