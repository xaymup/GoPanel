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
    "archive/zip"
    "io/fs"
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

type FileUpdateRequest struct {
    FilePath string `json:"file_path"`
    Content  string `json:"content"`
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

    info, _ := os.Stat(req.Source)


	if info.IsDir() {
        if err := copyDir(req.Source, req.Destination); err != nil {
            http.Error(w, fmt.Sprintf("Error copying file: %v", err), http.StatusInternalServerError)
            return
        }
		// Source is a directory, copy recursively
	} else {
        if err := copyFile(req.Source, req.Destination); err != nil {
            http.Error(w, fmt.Sprintf("Error copying file: %v", err), http.StatusInternalServerError)
            return
        }
		// Source is a file, copy file
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

func copyDir(source string, destination string) error {
    destDirPath := generateFileName(destination)

	err := os.MkdirAll(destDirPath, os.ModePerm)
	if err != nil {
		return err
	}

	return filepath.WalkDir(source, func(srcPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, srcPath)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destDirPath, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, os.ModePerm)
		} else {
			return copyFile(srcPath, destPath)
		}
	})
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

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the file path from the query parameters
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Try to remove the file
	err := os.RemoveAll(filePath)
	if err != nil {
		// If an error occurred, send an appropriate response
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		}
		return
	}

	// Send a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File deleted successfully: %s", filePath)
}

func CompressFile (w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the file/folder path from the query parameters
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "File or folder path is required", http.StatusBadRequest)
		return
	}

	// Ensure the path is absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve absolute path: %v", err), http.StatusInternalServerError)
		return
	}

	// Extract the base name of the file/folder (for creating the zip file name)
	baseName := filepath.Base(absPath)
	zipFileName := baseName + ".zip"
	zipFilePath := filepath.Join(filepath.Dir(absPath), zipFileName)

	// Create the zip file
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create zip file: %v", err), http.StatusInternalServerError)
		return
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk the path to compress
	err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the correct zip file path, relative to the root folder
		relPath := strings.TrimPrefix(filePath, filepath.Dir(absPath))
		if relPath == "" {
			relPath = filepath.Base(absPath) // Include the base name if it's the root folder
		} else if strings.HasPrefix(relPath, string(os.PathSeparator)) {
			relPath = relPath[1:] // Remove leading separator if present
		}

		// If the file is a directory, just add it to the zip without compression
		if info.IsDir() {
			_, err := zipWriter.Create(relPath + "/")
			if err != nil {
				return err
			}
			return nil
		}

		// If the file is not a directory, add it to the zip
		zipFileEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Open the file to read its contents
		fileToCompress, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fileToCompress.Close()

		// Copy the file contents to the zip
		_, err = io.Copy(zipFileEntry, fileToCompress)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to compress folder or file: %v", err), http.StatusInternalServerError)
		return
	}

	// Send a success message with the zip file path
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully compressed to: %s", zipFilePath)
}


func ExtractFile(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the ZIP file path from the query parameters
	zipPath := r.URL.Query().Get("path")
	if zipPath == "" {
		http.Error(w, "ZIP file path is required", http.StatusBadRequest)
		return
	}

	// Ensure the path is absolute
	absZipPath, err := filepath.Abs(zipPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve absolute path: %v", err), http.StatusInternalServerError)
		return
	}

	// Open the ZIP file
	zipFile, err := os.Open(absZipPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open ZIP file: %v", err), http.StatusInternalServerError)
		return
	}
	defer zipFile.Close()

    // Get the file info for the ZIP file
	zipFileInfo, err := zipFile.Stat()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file info: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a zip reader
	zipReader, err := zip.NewReader(zipFile, zipFileInfo.Size())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create zip reader: %v", err), http.StatusInternalServerError)
		return
	}

	// Create the extraction directory (same directory as the ZIP file)
	extractDir := strings.TrimSuffix(absZipPath, filepath.Ext(absZipPath))
	err = os.MkdirAll(extractDir, os.ModePerm)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create extraction directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Iterate through each file in the ZIP archive
	for _, file := range zipReader.File {
		// Open the file inside the ZIP
		fileInZip, err := file.Open()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to open file in ZIP: %v", err), http.StatusInternalServerError)
			return
		}
		defer fileInZip.Close()

		// Determine the path to extract to
		filePath := filepath.Join(extractDir, file.Name)

		// Create the necessary directories
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
				return
			}
			continue
		}

		// Create the file to extract to
		destFile, err := os.Create(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
			return
		}
		defer destFile.Close()

		// Copy the file contents
		_, err = io.Copy(destFile, fileInZip)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to copy file contents: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Send a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully extracted ZIP file to: %s", extractDir)
}


// getFileContents reads and returns the contents of a file
func GetFile(w http.ResponseWriter, r *http.Request) {
	// Query parameter to get the file path
	filePath := r.URL.Query().Get("path")

	// Validate file path input
	if filePath == "" {
		http.Error(w, "File path is missing", http.StatusBadRequest)
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Read the file contents
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Send the file contents as the HTTP response
	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}


func UpdateFile(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var request FileUpdateRequest
    err := json.NewDecoder(r.Body).Decode(&request)
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    if request.FilePath == "" {
        http.Error(w, "File path cannot be empty", http.StatusBadRequest)
        return
    }

    err = ioutil.WriteFile(request.FilePath, []byte(request.Content), 0644)
    if err != nil {
        http.Error(w, "Failed to write file", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("File updated successfully"))
}