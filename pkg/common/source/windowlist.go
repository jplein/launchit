package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/jplein/launchit/pkg/common/desktop"
	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/server"
	"github.com/jplein/launchit/pkg/common/source/niri"
)

type WindowList struct{}

const (
	windowListSourceName = "windows"
	windowListSourceType = "Window"
	windowListPrefix     = "window"
)

func listFromServer() ([]niri.WindowDescription, error) {
	url := fmt.Sprintf("http://127.0.0.1:%s/api/v1/windows", server.Port)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting windows from server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error getting windows from server: status %d: %s", resp.StatusCode, string(body))
	}

	var windows []niri.WindowDescription
	if err := json.NewDecoder(resp.Body).Decode(&windows); err != nil {
		return nil, fmt.Errorf("error parsing windows JSON: %w", err)
	}

	return windows, nil
}

func (w *WindowList) List() ([]Entry, error) {
	// Try to get windows from the server first
	windows, err := listFromServer()
	if err != nil {
		// If server is unavailable, fall back to direct niri command
		logger.Log("error getting windows from server, falling back to direct niri command: %v\n", err)
		windows, err = niri.ListWindows()
		if err != nil {
			return nil, err
		}
	}

	entries := make([]Entry, 0)

	for _, window := range windows {
		desktopEntry, err := desktop.Get(window.AppID)
		if err != nil {
			logger.Log("error getting desktop entry %s: %v\n", window.AppID, err)
		}

		icon := ""
		override := getIconOverride(window.AppID)

		switch {
		case override != "":
			icon = override
		case desktopEntry != nil:
			icon = desktopEntry.Icon
		}

		// Attempt to get the name from the desktop entry, if we can find it. Otherwise, leave it blank.
		name := window.AppID
		if desktopEntry != nil {
			name = desktopEntry.Name
		}

		entry := Entry{
			Description: fmt.Sprintf("%s (%s)", window.Title, name),
			ID:          windowListPrefix + ":" + fmt.Sprintf("%d", window.ID),
			Icon:        icon,
			Type:        windowListSourceType,
			Hidden:      window.AppID,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (w *WindowList) Name() string {
	return windowListSourceName
}

func (w *WindowList) Handle(entry Entry) error {
	id := entry.ID

	if !strings.HasPrefix(id, windowListPrefix+":") {
		return fmt.Errorf("not a Niri window: %s", id)
	}

	windowId := id[len(windowListPrefix)+1:]
	if windowId == "" {
		return fmt.Errorf("not a valid ID: window ID is empty")
	}

	for i, c := range windowId {
		if c < '0' && c > '9' {
			return fmt.Errorf("not a valid window ID: character %d is not a digit: '%c'", i, c)
		}
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("niri", "msg", "action", "focus-window", "--id", windowId)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stdout.Len() > 0 {
			logger.Log("niri focus-window stdout: %s\n", stdout.String())
		}
		if stderr.Len() > 0 {
			logger.Log("niri focus-window stderr: %s\n", stderr.String())
		}
		return fmt.Errorf("error switching to window %s: %w", windowId, err)
	}

	if stdout.Len() > 0 {
		logger.Log("niri focus-window stdout: %s\n", stdout.String())
	}
	if stderr.Len() > 0 {
		logger.Log("niri focus-window stderr: %s\n", stderr.String())
	}

	return nil
}

func (w *WindowList) Prefix() string {
	return windowListPrefix
}

// TODO: This should be in a configuration file that can be edited without rebuilding the application
var overrides = map[string]string{
	"google-chrome": "com.google.Chrome",
	"Code":          "vscode",
}

func getIconOverride(appID string) string {
	return overrides[appID]
}
