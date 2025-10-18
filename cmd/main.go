package main

import (
	"fmt"
	"os"

	"github.com/jplein/launchit/pkg/common/launcher"
	"github.com/jplein/launchit/pkg/common/source"
)

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

	entries, err := l.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting list of entries: %v", err)
		os.Exit(1)
	}

	fmt.Printf("len(entries): %d\n", len(entries))

	err = l.Write(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing entries: %v", err)
		os.Exit(1)
	}
}
