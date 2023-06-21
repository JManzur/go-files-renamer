package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	// Define command-line flags
	folderPath := flag.String("folder", "", "Path to the folder containing the files")
	maxGoroutines := flag.Int("max-goroutines", 20, "Maximum number of simultaneous goroutines")
	logFilePath := flag.String("log-file", "", "Path to the log file")
	verbose := flag.Bool("v", false, "Enable verbose mode (outputs to console)")
	flag.Parse()

	// Check if the folder path is provided
	if *folderPath == "" {
		log.Fatal("Please provide the path to the folder using the '-folder' flag")
	}

	// Set log output to a file
	logFile := *logFilePath
	if logFile == "" {
		logFile = "renamer.log"
	}

	// Check if the log file already exists
	if _, err := os.Stat(logFile); err == nil {
		// Archive the existing log file by appending a timestamp
		timestamp := time.Now().Format("20060102150405")
		archivedLogFile := fmt.Sprintf("%s_%s", logFile, timestamp)
		if err := os.Rename(logFile, archivedLogFile); err != nil {
			log.Fatal(err)
		}
		log.Printf("Archived existing log file: %s -> %s", logFile, archivedLogFile)
	}

	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a multi-writer to write to both log file and console if verbose mode is enabled
	var writers []io.Writer
	writers = append(writers, file)
	if *verbose {
		writers = append(writers, os.Stdout)
	}
	multiWriter := io.MultiWriter(writers...)

	// Set the log output to the multi-writer
	log.SetOutput(multiWriter)

	// Create a semaphore with the maximum number of goroutines
	semaphore := make(chan struct{}, *maxGoroutines)

	// Get a list of files in the folder
	files, err := ioutil.ReadDir(*folderPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create a wait group to ensure all goroutines complete
	var wg sync.WaitGroup
	wg.Add(len(files))

	// Iterate over the files and launch goroutines to rename them
	for _, file := range files {
		// Acquire a semaphore slot
		semaphore <- struct{}{}

		go func(file os.FileInfo) {
			defer func() {
				// Release the semaphore slot
				<-semaphore
				wg.Done()
			}()

			// Skip directories
			if file.IsDir() {
				return
			}

			// Construct the old and new file paths
			oldPath := filepath.Join(*folderPath, file.Name())
			newPath := filepath.Join(*folderPath, strings.ToLower(file.Name()))

			// Get the original file permissions
			oldPermissions := file.Mode().String()

			// Get the original file ownership
			oldUID, oldGID, err := getFileOwnership(file)
			if err != nil {
				log.Printf("Failed to retrieve ownership for file: %s", oldPath)
			}

			// Rename the file to lowercase
			err = os.Rename(oldPath, newPath)
			if err != nil {
				log.Printf("Failed to rename file: %s", oldPath)
			} else {
				// Get the updated file permissions
				newFile, err := os.Stat(newPath)
				if err != nil {
					log.Printf("Failed to retrieve updated permissions for file: %s", newPath)
					return
				}
				newPermissions := newFile.Mode().String()

				// Set the file ownership
				err = setFileOwnership(newPath, oldUID, oldGID)
				if err != nil {
					log.Printf("Failed to set ownership for file: %s", newPath)
				}

				// Log the file renaming and permissions
				log.Printf("Renamed file: %s to %s", oldPath, newPath)
				log.Printf("File: %s - Permissions - Before: %s, After: %s", newPath, oldPermissions, newPermissions)
				log.Printf("File: %s - Ownership - UID: %d, GID: %d", newPath, oldUID, oldGID)
			}
		}(file)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	log.Println("All files renamed to lowercase.")
}

// getFileOwnership returns the user ID (UID) and group ID (GID) of a file
func getFileOwnership(file os.FileInfo) (uint32, uint32, error) {
	fileSys := file.Sys()
	if stat, ok := fileSys.(*syscall.Stat_t); ok {
		return stat.Uid, stat.Gid, nil
	}
	return 0, 0, fmt.Errorf("failed to retrieve file ownership")
}

// setFileOwnership sets the user ID (UID) and group ID (GID) of a file
func setFileOwnership(file string, uid, gid uint32) error {
	if err := os.Lchown(file, int(uid), int(gid)); err != nil {
		return err
	}
	return nil
}
