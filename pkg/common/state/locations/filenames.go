package locations

import (
	"fmt"
	"path"
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
