package source

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jplein/launchit/pkg/common/desktop"
	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/niri"
	"github.com/jplein/launchit/pkg/overrides"
)

const (
	windowListSourceName = "windows"
	windowListSourceType = "Window"
	windowListPrefix     = "window"
)

type WindowList struct{}

func (w *WindowList) List() ([]Entry, error) {
	windows, err := niri.ListWindows(true)
	if err != nil {
		return nil, fmt.Errorf("error getting window list: %w", err)
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

func getIconOverride(appID string) string {
	o, err := overrides.ByAppID(appID)
	if err != nil {
		logger.Log("error getting overrides for app ID %s: %w", appID, err)
		return ""
	}

	return o.Icon
}
