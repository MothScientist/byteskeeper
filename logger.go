package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
)

func setupLogs(sessionUuidString string) *os.File {
	err := createDirectory(sessionUuidString)
	if err != nil {
		log.Fatalf("Error creating logs directory: %v;", err)
	}

	filePath := fmt.Sprintf("%s/main.log", sessionUuidString)
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o0666)
	if err != nil {
		log.Fatalf("Error creating/opening file .log: %v;", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return logFile
}

// createDirectory creates a logs directory in the current working directory if it does not exist.
func createDirectory(sessionUuidString string) error {
	// Get current working directory
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Forming the full path to the logs directory
	logsDir := filepath.Join(dir, sessionUuidString)

	// Check for directory existence
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		err := os.Mkdir(logsDir, 0o0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		// Handling other Stat errors
		return err
	}
	return nil
}

func logPanic() {
	if r := recover(); r != nil {
		log.Printf(
			"PANIC: %v\nStack trace:\n%s;",
			r,
			string(debug.Stack()),
		)
		os.Exit(1)
	}
}
