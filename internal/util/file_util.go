package util

import (
    "os"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"net/http"
	"time"
	"os/user"
	"syscall"
	"encoding/json"
)

type FileDetail struct {
	Name       string    `json:"name"`
	Size       string     `json:"size"`
	Type       string    `json:"type"`
	Modified   string `json:"modified"`
	Owner      string    `json:"owner"`
	Permissions string   `json:"permissions"`
}



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

func detectMimeType(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    buffer := make([]byte, 512)
    _, err = file.Read(buffer)
    if err != nil {
        return "", err
    }

    mimeType := http.DetectContentType(buffer)
    return mimeType, nil
}

func ListFiles (w http.ResponseWriter, r *http.Request) {
	// Parse the "path" query parameter
	queryPath := r.URL.Query().Get("path")
	if queryPath == "" {
		http.Error(w, "Path query parameter is required", http.StatusBadRequest)
		return
	}

	// Get the absolute path
	absPath, err := filepath.Abs(queryPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Open the directory
	dir, err := os.Open(absPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to open directory: %v", err), http.StatusInternalServerError)
		return
	}
	defer dir.Close()

	// Read the directory contents
	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read directory: %v", err), http.StatusInternalServerError)
		return
	}

	var files []FileDetail

	for _, fileInfo := range fileInfos {
		filePath := filepath.Join(absPath, fileInfo.Name())

		// Get the file owner
		stat := fileInfo.Sys().(*syscall.Stat_t)
		uid := stat.Uid
		fileOwner, err := user.LookupId(fmt.Sprint(uid))
		if err != nil {
			fileOwner = &user.User{Username: "unknown"}
		}

		// Get file permissions
		permissions := fileInfo.Mode().Perm()

        var fileType string
        if fileInfo.IsDir() {
            fileType = "directory"
        } else {
            mimeType, err := detectMimeType(filePath)
            if err != nil {
                log.Println("Error detecting MIME type:", err)
                fileType = "unknown"
            } else {
                fileType = mimeType
            }
        }


		// Append file or directory details to the list
		files = append(files, FileDetail{
			Name:        fileInfo.Name(),
			Size:        formatSize(fileInfo.Size()),
			Type:        fileType,
			Modified:    fileInfo.ModTime().Format(time.Stamp),
			Owner:       fileOwner.Username,
			Permissions: fmt.Sprintf("%#o", permissions),
		})
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")


	// Encode the response as JSON and send it
	if err := json.NewEncoder(w).Encode(files); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func formatSize(bytes int64) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
        TB = GB * 1024
    )

    switch {
    case bytes >= TB:
        return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
    case bytes >= GB:
        return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
    case bytes >= MB:
        return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
    case bytes >= KB:
        return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
    default:
        return fmt.Sprintf("%d B", bytes)
    }
}

func calculateDirSize(path string) (int64, error) {
    var totalSize int64

    err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            totalSize += info.Size()
        }
        return nil
    })

    if err != nil {
        return 0, err
    }
    return totalSize, nil
}