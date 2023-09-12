package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	// Perform the recursive renaming starting from the specified folder path
	err = renameToLowerCaseRecursive(*folderPath, *folderPath, semaphore)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("All files and folders under the specified folder path renamed to lowercase.")
}

func renameToLowerCaseRecursive(rootPath, currentPath string, semaphore chan struct{}) error {
	// Acquire a semaphore slot
	semaphore <- struct{}{}
	defer func() {
		// Release the semaphore slot
		<-semaphore
	}()

	// Rename the current folder to lowercase
	lowerPath := filepath.Join(rootPath, strings.ToLower(currentPath[len(rootPath):]))
	if currentPath != lowerPath {
		if err := os.Rename(currentPath, lowerPath); err != nil {
			return err
		}
		log.Printf("Renamed folder: %s to %s", currentPath, lowerPath)
	}

	// Get a list of files and subfolders in the current folder
	entries, err := readDir(currentPath)
	if err != nil {
		return err
	}

	// Create a wait group to ensure all goroutines complete
	var wg sync.WaitGroup

	for _, entry := range entries {
		entryPath := filepath.Join(currentPath, entry.Name())

		if entry.IsDir() {
			// If it's a subfolder, recursively rename it
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := renameToLowerCaseRecursive(rootPath, entryPath, semaphore); err != nil {
					log.Printf("Error renaming folder: %s", entryPath)
				}
			}()
		} else {
			// If it's a file, rename it to lowercase
			lowerEntryPath := filepath.Join(currentPath, strings.ToLower(entry.Name()))
			if entryPath != lowerEntryPath {
				if err := os.Rename(entryPath, lowerEntryPath); err != nil {
					log.Printf("Failed to rename file: %s", entryPath)
				} else {
					log.Printf("Renamed file: %s to %s", entryPath, lowerEntryPath)
				}
			}
		}
	}

	// Wait for all subfolder renaming goroutines to complete
	wg.Wait()

	return nil
}

// readDir is a replacement for ioutil.ReadDir
func readDir(dirname string) ([]os.FileInfo, error) {
	file, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	names, err := file.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	var list []os.FileInfo
	for _, name := range names {
		info, err := os.Stat(filepath.Join(dirname, name))
		if err != nil {
			return nil, err
		}
		list = append(list, info)
	}

	return list, nil
}
