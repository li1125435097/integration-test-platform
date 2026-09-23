package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectRoot finds the directory containing go.mod by walking up from cwd.
func ProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}

// ExecutableDir returns the directory of the current executable.
func ExecutableDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

// ResolveConfigDir picks config directory: flag/env > exe/config > project/config.
func ResolveConfigDir(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	if exeDir, err := ExecutableDir(); err == nil {
		candidate := filepath.Join(exeDir, "config")
		if _, err := os.Stat(filepath.Join(candidate, "menu.json")); err == nil {
			return candidate, nil
		}
	}
	root, err := ProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "config"), nil
}

// DevWebDir returns repository root web/ for development (disk serving).
func DevWebDir() (string, error) {
	root, err := ProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "web"), nil
}
