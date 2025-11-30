package logger

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/jplein/launchit/pkg/common/state/locations"
)

var (
	logFile     string
	logFH       *os.File
	subcommand  string
	initialized bool
)

// Init initializes the logger with the specified subcommand.
// This determines which log file to use (e.g., launchit-write.log, launchit-read.log).
func Init(subcmd string) {
	subcommand = subcmd
	initialized = true
}

func Log(format string, a ...any) {
	// Add timestamp to the log message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, a...)
	timestampedMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	fmt.Fprintf(os.Stderr, timestampedMessage)

	fh, err := getLogFilehandle()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting log file: %v", err)
		return
	}

	fmt.Fprintf(fh, timestampedMessage)
}

func getLogFile() (string, error) {
	if logFile != "" {
		return logFile, nil
	}

	var err error
	if initialized && subcommand != "" {
		logFile, err = locations.LogFilenameForSubcommand(subcommand)
	} else {
		logFile, err = locations.LogFilename()
	}

	if err != nil {
		return "", fmt.Errorf("error getting log file: %w", err)
	}

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

	logDir := path.Dir(file)
	if err = os.MkdirAll(logDir, 0o744); err != nil {
		return nil, fmt.Errorf("error opening log file %s: error creating directory %s: %w",
			file, logDir, err)
	}

	fh, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}

	logFH = fh
	return fh, nil
}
