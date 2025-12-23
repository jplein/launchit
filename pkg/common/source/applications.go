package source

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/jplein/launchit/pkg/common/desktop"
	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/niri"
	"github.com/jplein/launchit/pkg/overrides"
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

	windows, err := niri.ListWindows(true)
	if err != nil {
		logger.Log("error collecting window list: %w", err)
		windows = []niri.WindowDescription{}
	}

	entries := make([]Entry, 0)
	for _, app := range apps {
		desc := app.Name
		window := getWindow(app, windows)

		if window != nil {
			desc = fmt.Sprintf("• %s", desc)
		}

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

	app, err := desktop.FromFile(filename)
	if err != nil {
		return fmt.Errorf("error reading desktop entry from file %s: %w", filename, err)
	}

	if err = a.exec(app); err != nil {
		return fmt.Errorf("error running application: %w", err)
	}

	return nil
}

func (a *Applications) Prefix() string {
	return idPrefix
}

func (a *Applications) exec(app desktop.App) error {
	sh, err := exec.LookPath("sh")
	if err != nil {
		return fmt.Errorf("error starting application: could not find sh in the PATH")
	}

	if app.Exec == "" {
		return fmt.Errorf("error starting application from file %s: Exec entry is missing or blank", app.Filename)
	}

	if app.Path != "" {
		if err := os.Chdir(app.Path); err != nil {
			return fmt.Errorf("error starting application: error setting working directory '%s': %w", app.Path, err)
		}
	}

	env := os.Environ()

	args := []string{sh, "-c", app.Exec}
	err = syscall.Exec(sh, args, env)
	if err != nil {
		return fmt.Errorf("error starting application: error executing %s with arguments %v: %w", sh, args, err)
	}

	return nil
}

// Returns the most recently accessed open window for the application, or nil if
// there is no such window
//
// entry: An application entry
//
// windows: The list of open windows, with the most recently accessed windows at
// the beginning of the list, as returned by niri.ListWindows()
func getWindow(app desktop.App, windows []niri.WindowDescription) *niri.WindowDescription {
	id := app.ID

	or, err := overrides.ByAppID(app.ID)
	if err != nil {
		logger.Log("error getting window ID for app %s: %w", app.ID, err)
	}

	if or != nil {
		id = or.WindowAppID
	}

	for _, window := range windows {
		if window.AppID == id {
			return &window
		}
	}

	return nil
}
