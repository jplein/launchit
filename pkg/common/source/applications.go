package source

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/jplein/launchit/pkg/common/desktop"
)

type Applications struct{}

const (
	idPrefix   = "app"
	sourceName = "applications"
)

func (a *Applications) List() ([]Entry, error) {
	apps, err := desktop.List()
	if err != nil {
		return nil, fmt.Errorf("error listing applications: %w", err)
	}

	entries := make([]Entry, 0)
	for _, app := range apps {
		entry := Entry{
			Description: app.Name,
			Icon:        app.Icon,
			ID:          idPrefix + ":" + app.Filename,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (a *Applications) Name() string {
	return sourceName
}

func (a *Applications) Handle(entry Entry) error {
	id := entry.ID
	if !strings.HasPrefix(id, idPrefix+":") {
		return fmt.Errorf("not an application: %s", id)
	}

	filename := id[len(idPrefix)+1:]
	if filename == "" {
		return fmt.Errorf("not a valid ID: filename is empty: %s", id)
	}

	cmd := exec.Command("gio", "launch", filename)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch application %s: %w", filename, err)
	}

	return nil
}

func (a *Applications) Prefix() string {
	return idPrefix
}
