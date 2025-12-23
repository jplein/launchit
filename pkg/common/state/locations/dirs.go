package locations

import (
	"fmt"
	"os"
	"path"
)

const (
	// relative to the home directory
	defaultXDGDataHome   = ".local/state"
	defaultXDGConfigHome = ".config"
	appName              = "launchit"
)

func StateDirectory() (string, error) {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting state directory location: %w", err)
		}

		xdgDataHome = path.Join(home, defaultXDGDataHome)
	}

	return path.Join(xdgDataHome, appName), nil
}

func ConfigDirectory() (string, error) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting config directory location: %s", err)
		}

		xdgConfigHome = path.Join(home, defaultXDGConfigHome)
	}

	return path.Join(xdgConfigHome, appName), nil
}
