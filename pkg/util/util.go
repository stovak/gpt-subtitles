package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	cwd := filepath.Dir(filename)

	// Find the root of the repository.
	repoRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err == nil {
			break
		}

		repoRoot = filepath.Dir(repoRoot)

		if repoRoot == "/" {
			fmt.Println("Could not find go.mod directory.")
			return ""
		}
	}
	return repoRoot
}

func TrimSlice(slice []string) []string {
	var toReturn []string
	for i, s := range slice {
		// if the string is not empty, add it to the slice
		if strings.Trim(s, "") != "" {
			toReturn = append(toReturn, slice[i])
			break
		}

	}
	return toReturn
}
