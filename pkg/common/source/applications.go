package source

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/jplein/launchit/pkg/common/desktop"
	"gopkg.in/ini.v1"
)

type Applications struct{}

const (
	idPrefix      = "app"
	appSourceName = "applications"
	appSourceType = "Application"
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
			Type:        appSourceType,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (a *Applications) Name() string {
	return appSourceName
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

	desktopFileEntry, err := ini.Load(filename)
	if err != nil {
		return fmt.Errorf("error parsing %s as an ini file: %w", err)
	}

	desktopSection := desktopFileEntry.Section("Desktop Entry")

	cmd := desktopSection.Key("Exec").String()
	if cmd == "" {
		return fmt.Errorf("error getting command from %s: no Exec line found", filename)
	}

	// Remove any positional arguments (like %U)
	fieldCodes := []string{"%f", "%F", "%u", "%U", "%i", "%c", "%k", "%d", "%D", "%n", "%N", "%v", "%m"}
	for _, code := range fieldCodes {
		cmd = strings.ReplaceAll(cmd, code, "")
	}
	cmd = strings.TrimSpace(cmd)

	sh, err := exec.LookPath("sh")
	if err != nil {
		return fmt.Errorf("error starting application: could not find sh in the PATH")
	}

	if path := desktopSection.Key("Path").String(); path != "" {
		if err := os.Chdir(path); err != nil {
			return fmt.Errorf("error starting application: error setting working directory '%s': %w", path, err)
		}
	}

	env := os.Environ()

	args := []string{sh, "-c", cmd}
	err = syscall.Exec(sh, args, env)
	if err != nil {
		return fmt.Errorf("error starting application: error executing %s with arguments %v: %w", sh, args, err)
	}

	return nil
}

func (a *Applications) Prefix() string {
	return idPrefix
}
