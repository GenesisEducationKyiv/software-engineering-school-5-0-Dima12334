package config

import (
	"log"
	"os"
	"path/filepath"
)

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	for {
		// Check if go.mod exists in this directory
		modFile := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modFile); err == nil {
			return dir
		}

		// Move one directory up
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, stop searching
			log.Fatal("could not find go.mod; project root not found")
		}
		dir = parent
	}
}

// GetOriginalPath constructs the absolute path to the original file or folder
// based on the project root directory.
func GetOriginalPath(objectPath string) string {
	rootDir := findProjectRoot()
	return filepath.Join(rootDir, objectPath)
}
