package main

import (
	"fmt"
	"os"

	"github.com/jplein/launchit/pkg/common/launcher"
	"github.com/jplein/launchit/pkg/common/source"
)

// TODO:
// - Read from stdin, if available
// - The second element in the single tab-separated line that should be available will be the ID
// - Use the prefix of the ID to route to the right source
// - Call Act() on the right source

func main() {
	sources, err := source.DefaultSourceSet()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting launcher: %v", err)
		os.Exit(1)
	}

	l, err := launcher.NewLauncher(*sources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting launcher: %v", err)
		os.Exit(1)
	}

	err = l.Write(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing entries: %v", err)
		os.Exit(1)
	}
}
