package main

import (
	"log"
	"os"
	"path/filepath"
)

func validatePath(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	var d []byte
	fp := filepath.Join(path, ".test_path_bytekeep")
	if err := os.WriteFile(fp, d, 0644); err == nil {
		os.Remove(fp)
		return true
	} else {
		log.Println("[ERROR] Invalid path; ", err)
	}
	return false
}