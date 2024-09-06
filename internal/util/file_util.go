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
    "io"
    "strings"
)

type FileDetail struct {
	Name       string    `json:"name"`
	Size       string     `json:"size"`
	Type       string    `json:"type"`
	Modified   string `json:"modified"`
	Owner      string    `json:"owner"`
	Permissions string   `json:"permissions"`
}

type FileRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}


type DownloadRequest struct {
	FilePath string `json:"filePath"`
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

func UploadFile(w http.ResponseWriter, r *http.Request) {
    // Restricting the request method to POST
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Get the path parameter from the query string
    path := r.URL.Query().Get("path")
    if path == "" {
        http.Error(w, "Path parameter is missing", http.StatusBadRequest)
        return
    }

    // Create the destination directory if it doesn't exist
    if err := os.MkdirAll(path, 0755); err != nil {
        http.Error(w, "Could not create directory", http.StatusInternalServerError)
        return
    }

    // Create a multipart reader to handle file uploads without memory limits
    reader, err := r.MultipartReader()
    if err != nil {
        http.Error(w, "Could not create multipart reader", http.StatusBadRequest)
        return
    }

    // Iterate over each part of the multipart request
    for {
        part, err := reader.NextPart()
        if err == io.EOF {
            break // No more parts
        }
        if err != nil {
            http.Error(w, "Error reading file part", http.StatusInternalServerError)
            return
        }

        // Skip form fields, process only the file part
        if part.FileName() == "" {
            continue
        }

        // Create destination file in the specified path
        filePath := filepath.Join(path, part.FileName())
        dst, err := os.Create(filePath)
        if err != nil {
            http.Error(w, "Could not save file", http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        // Copy the uploaded file data to the destination file
        _, err = io.Copy(dst, part)
        if err != nil {
            http.Error(w, "Could not copy file", http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "File uploaded successfully: %s\n", filePath)
    }
}


func RenameFile(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the incoming request body
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Check if both paths are provided
	if req.Source == "" || req.Destination == "" {
		http.Error(w, "Both oldPath and newPath must be provided", http.StatusBadRequest)
		return
	}

	// Clean up and resolve the new file path
	newFullPath := filepath.Clean(req.Destination)
    newFullPath = generateFileName(newFullPath)

	// Rename the file
	err := os.Rename(req.Source, newFullPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error renaming file: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("File renamed from %s to %s", req.Source, req.Destination)))
}

func CopyFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var req FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Source == "" || req.Destination == "" {
		http.Error(w, "Both source and destination paths are required", http.StatusBadRequest)
		return
	}

	// Copy file
	if err := copyFile(req.Source, req.Destination); err != nil {
		http.Error(w, fmt.Sprintf("Error copying file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("File copied from %s to %s", req.Source, req.Destination)))
}

// Helper function to copy the file content
func copyFile(source string, destination string) error {
	// Generate the correct destination file name with (1), (2), etc. if needed
	destFilePath := generateFileName(destination)

	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Ensure the destination directory exists
	destFile, err := os.Create(filepath.Clean(destFilePath))
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the file content
	_, err = io.Copy(destFile, srcFile)
	return err
}


func generateFileName(destination string) string {
	// Check if the file exists
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		// File does not exist, return the original destination
		return destination
	}

	// Split the file name and extension
	dir := filepath.Dir(destination)
	ext := filepath.Ext(destination)
	baseName := strings.TrimSuffix(filepath.Base(destination), ext)

	// Try adding (1), (2), etc. until a non-existing file is found
	for i := 1; ; i++ {
		newFileName := fmt.Sprintf("%s(%d)%s", baseName, i, ext)
		newFilePath := filepath.Join(dir, newFileName)

		if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
			// If the file does not exist, return the new file name
			return newFilePath
		}
	}
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the file path from the query parameters
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Serve the file as a downloadable attachment
	fileName := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	http.ServeFile(w, r, filePath)
}