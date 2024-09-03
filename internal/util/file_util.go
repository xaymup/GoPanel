package util

import (
    "os"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
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

func RemoveAllFiles(dir string) error {
    // Read all files and directories in the specified directory
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        return fmt.Errorf("could not read directory: %v", err)
    }

    // Iterate over each file/directory found
    for _, file := range files {
        // Construct the full file path
        filePath := filepath.Join(dir, file.Name())

        // Remove the file or directory
        err := os.RemoveAll(filePath)
        if err != nil {
            return fmt.Errorf("could not remove file: %v", err)
        }
    }

    return nil
}