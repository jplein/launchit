package main

import (
	"errors"
	"flag"
	"io"
	"os"

	"github.com/jplein/launchit/pkg/common/launcher"
	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/source"
	"github.com/jplein/launchit/pkg/common/state"
)

// TODO:
// - Read from stdin, if available
// - The second element in the single tab-separated line that should be available will be the ID
// - Use the prefix of the ID to route to the right source
// - Call Act() on the right source

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		writeEntries(args)
		return
	}

	subcommand := "write"
	if len(args) >= 1 {
		subcommand = args[0]
	}

	switch subcommand {
	case "read":
		handleInput()
	case "write":
		writeEntries(args[1:])
	default:
		logger.Log("unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func writeEntries(args []string) {
	fs := flag.NewFlagSet("write", flag.ExitOnError)
	src := fs.String("source", "", "source to pull entries from")
	fs.Parse(args)

	sources, err := source.DefaultSourceSet()
	if err != nil {
		logger.Log("error getting launcher: %v", err)
		os.Exit(1)
	}

	logger.Log("debug: *src: %s\n", *src)

	if *src != "" {
		foundSource := false

		for _, s := range sources.Sources {
			if s.Name() == *src {
				sources, err = source.NewSourceSet([]source.Source{s})
				if err != nil {
					logger.Log("error getting launcher: %v\n", err)
					os.Exit(1)
				}
				foundSource = true
			}
		}

		if !foundSource {
			logger.Log("error getting launcher: no source with name %s\n", *src)
		}
	}

	l, err := launcher.NewLauncher(*sources)
	if err != nil {
		logger.Log("error getting launcher: %v", err)
		os.Exit(1)
	}

	err = l.Write(os.Stdout)
	if err != nil {
		logger.Log("error writing entries: %v", err)
		os.Exit(1)
	}
}

func handleInput() {
	input, err := readFromSTDIN()
	if err != nil {
		logger.Log("error reading from standard input: %v\n", err)
		return
	}

	if input == "" {
		logger.Log("no input from standard input\n")
		return
	}

	entry, err := source.EntryFromString(input)
	if err != nil {
		logger.Log("%s\n", err.Error())
		os.Exit(1)
	}

	sources, err := source.DefaultSourceSet()
	if err != nil {
		logger.Log("error getting launcher: %v", err)
		os.Exit(1)
	}

	err = state.Add(entry.ID)
	if err != nil {
		logger.Log("error writing recent entry %s: %v", entry.ID, err)
	}

	err = sources.Handle(entry)
	if err != nil {
		logger.Log("error handling entry ('%s', '%s'): %v", entry.Description, entry.ID, err)
		os.Exit(1)
	}
}

// Don't read more than this many bytes from stdin - we're expecting to get one line from fzf, fuzzel, etc.
const bufSize = 1024 * 1024

func readFromSTDIN() (string, error) {
	buf := make([]byte, bufSize)

	logger.Log("about to read from stdin\n")
	_, err := os.Stdin.Read(buf)
	logger.Log("done reading from stdin\n")

	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return string(buf), nil
}
