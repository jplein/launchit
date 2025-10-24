package logger

import (
	"fmt"
	"os"
	"path"
)

func Log(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)

	fh, err := getLogFilehandle()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting log file: %v", err)
		return
	}

	fmt.Fprintf(fh, format, a...)
}

var (
	logFile string
	logFH   *os.File
)

const (
	// relative to the home directory
	defaultXDGDataHome = ".local/state"
	appName            = "launchit"
	baseLogFilename    = "launchit.log"
)

func getLogFile() (string, error) {
	if logFile != "" {
		return logFile, nil
	}

	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting log file location: %w", err)
		}

		xdgDataHome = path.Join(home, defaultXDGDataHome)
	}

	logDirectory := path.Join(xdgDataHome, appName)

	fmt.Fprintf(os.Stderr, "logger: logDirectory: %s\n", logDirectory)
	if err := os.MkdirAll(logDirectory, 0o755); err != nil {
		return "", fmt.Errorf("error getting log file location: error creating directory: %w", err)
	}

	logFile = path.Join(logDirectory, baseLogFilename)
	fmt.Fprintf(os.Stderr, "logger: logFile: %s\n", logFile)
	return logFile, nil
}

func getLogFilehandle() (*os.File, error) {
	if logFH != nil {
		return logFH, nil
	}

	file, err := getLogFile()
	if err != nil {
		return nil, err
	}

	fh, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	logFH = fh
	return fh, nil
}
