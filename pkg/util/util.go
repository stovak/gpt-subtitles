package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
