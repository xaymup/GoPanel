package util

import (
    "os"
	"fmt"
	"io/ioutil"
)

func FileExists(path string) bool {
    _, err := os.Stat(path)
    return !os.IsNotExist(err)
}

func ReadSecretFromFile() (string, error) {
	filePath := "/etc/gopanel"
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", fmt.Errorf("error reading file: %w", err)
    }
    return string(data), nil
}