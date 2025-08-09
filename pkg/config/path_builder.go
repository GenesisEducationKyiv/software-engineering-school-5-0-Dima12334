package config

import (
	"log"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func FindProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	for {
		goWork := filepath.Join(dir, "go.work")

		if FileExists(goWork) {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal("could not find go.work; project root not found")
		}
		dir = parent
	}
}

// GetOriginalPath constructs the absolute path to the original file or folder
// based on the project root directory.
func GetOriginalPath(objectPath string) string {
	rootDir := FindProjectRoot()
	return filepath.Join(rootDir, objectPath)
}
