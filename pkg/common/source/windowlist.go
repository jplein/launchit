package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jplein/launchit/pkg/common/desktop"
	"github.com/jplein/launchit/pkg/common/logger"
)

type WindowList struct{}

const (
	windowListBufSize    = 16 * 1024 * 1024
	windowListSourceName = "windows"
	windowListPrefix     = "window"
)

type niriWindowDescription struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	AppID string `json:"app_id"`
}

func (w *WindowList) List() ([]Entry, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("niri", "msg", "--json", "windows")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error getting windows from Niri: %w", err)
	}

	listBytes := stdout.Bytes()

	windows := make([]niriWindowDescription, 0)
	if err := json.Unmarshal(listBytes, &windows); err != nil {
		logger.Log("niri window list JSON output:\n")
		logger.Log(string(listBytes))
		return nil, fmt.Errorf("error getting windows from Niri: error parsing JSON: %w", err)
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

		entry := Entry{
			Description: window.Title,
			ID:          windowListPrefix + ":" + fmt.Sprintf("%d", window.ID),
			Icon:        icon,
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
		return fmt.Errorf("error switching to window %d: %w", windowId, err)
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
