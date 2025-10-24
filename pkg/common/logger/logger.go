package logger

import (
	"fmt"
	"os"

	"github.com/jplein/launchit/pkg/common/state/locations"
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
	baseLogFilename = "launchit.log"
)

func getLogFile() (string, error) {
	if logFile != "" {
		return logFile, nil
	}

	logFile, err := locations.LogFilename()
	if err != nil {
		return "", fmt.Errorf("erorr getting log file: %w", err)
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

	fh, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	logFH = fh
	return fh, nil
}
