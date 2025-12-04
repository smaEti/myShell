package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	mode := info.Mode()
	return mode&0111 != 0
}

// findExecutable searches for an executable in PATH
func findExecutable(command string) string {
	pathString := os.Getenv("PATH")
	paths := strings.Split(pathString, string(os.PathListSeparator))

	for _, path := range paths {
		files, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		index := sort.Search(len(files), func(i int) bool {
			return files[i].Name() >= command
		})

		if index < len(files) && files[index].Name() == command {
			fullPath := filepath.Join(path, command)
			if isExecutable(fullPath) {
				return fullPath
			}
		}
	}

	return ""
}
