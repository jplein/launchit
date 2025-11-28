package locations

import (
	"fmt"
	"os"
	"path"
)

const (
	// relative to the home directory
	defaultXDGDataHome = ".local/state"
	appName            = "launchit"
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

const (
	baseLogFilename = "launchit.log"
)

func LogFilename() (string, error) {
	stateDirectory, err := StateDirectory()
	if err != nil {
		return "", err
	}

	return path.Join(stateDirectory, baseLogFilename), nil
}

func LogFilenameForSubcommand(subcommand string) (string, error) {
	stateDirectory, err := StateDirectory()
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("launchit-%s.log", subcommand)
	return path.Join(stateDirectory, filename), nil
}

const (
	baseRecentFilename = "recent.json"
)

func RecentFilename() (string, error) {
	stateDirectory, err := StateDirectory()
	if err != nil {
		return "", err
	}

	return path.Join(stateDirectory, baseRecentFilename), nil
}
