package locations

import (
	"fmt"
	"os"
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

type XDGDirectory string

const (
	XDGStateDir  XDGDirectory = "state"
	XDGConfigDir XDGDirectory = "config"

	WorldReadable         = os.FileMode(0o644)
	DefaultFilePermission = WorldReadable
)

// Get the path to configuration file
//
// dir: The ancestor directory, one of XDGStateDir or XDGConfigDir
//
// relPath: The path of the configuration file relative to the ancestor
// directory
func Get(dir XDGDirectory, relPath string) (string, error) {
	var baseDir string
	var err error

	switch dir {
	case XDGStateDir:
		baseDir, err = StateDirectory()
	case XDGConfigDir:
		baseDir, err = ConfigDirectory()
	default:
		err = fmt.Errorf("unknown dir %s, expected one of %s or %s", dir, XDGStateDir, XDGConfigDir)
	}

	if err != nil {
		return "", fmt.Errorf("error geting location: %w", err)
	}

	return path.Join(baseDir, relPath), nil
}

// Get the path to a configuration file, and initialize it if it is missing. If
// the file already exists, it is unmodified.
//
// dir: The ancestor directory, one of XDGStateDir or XDGConfigDir
//
// relPath: The path of the configuration file relative to the ancestor
// directory
//
// buf: The contents of the file to create if it is missing
//
// mode: The permissions to create it with if it is missing
func Initialize(dir XDGDirectory, relPath string, buf []byte, mode os.FileMode) (string, error) {
	file, err := Get(dir, relPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialize(): in err != nil branch")
		return "", fmt.Errorf("error initializing %s: %w", relPath, err)
	}

	// If the file already exists, don't modify it
	if _, err := os.Stat(file); err == nil {
		return file, nil
	}

	if err = os.WriteFile(file, buf, mode); err != nil {
		return "", fmt.Errorf("error initializing %s: %w", relPath, err)
	}

	return file, nil
}

const (
	baseLogFilename = "launchit.log"
)
