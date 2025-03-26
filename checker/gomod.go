package checker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/audi70r/uncle-bob/utilities/clog"
	"golang.org/x/mod/modfile"
)

// ModPath holds the Go module path from go.mod
var ModPath string

// LocateGoMod finds and parses the go.mod file
func LocateGoMod(targetPath string) {
	var err error

	clog.Info("Locating go.mod file...")
	ModPath, err = getModulePath(targetPath)

	if err != nil {
		clog.Error(err.Error())
		clog.Error("Could not find go.mod file. Please make sure you're in the root directory of a Go module.")
		os.Exit(1)
	}

	clog.Info(fmt.Sprintf("Found module: %s", ModPath))
}

// getModulePath extracts the module path from go.mod file
func getModulePath(targetPath string) (string, error) {
	// Ensure absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Read go.mod file
	gomodPath := filepath.Join(absPath, "go.mod")
	gomod, err := os.ReadFile(gomodPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Parse module path
	modulePath := modfile.ModulePath(gomod)
	if modulePath == "" {
		return "", fmt.Errorf("empty module path in go.mod")
	}

	return modulePath, nil
}
