package config

import (
	"log"
	"os"
	"path/filepath"
)

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	for {
		goMod := filepath.Join(dir, "go.mod")
		binDir := filepath.Join(dir, "bin")

		if fileExists(goMod) || dirExists(binDir) {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal("could not find go.mod or bin directory; project root not found")
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
