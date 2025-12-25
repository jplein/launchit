package source

import (
	"fmt"
	"strconv"
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
		appID := window.AppID

		or, err := overrides.ByWindowAppID(appID)
		if err != nil {
			logger.Log("error reading overrides: %v", err)
		}

		if or != nil {
			appID = or.AppID
		}

		desktopEntry, err := desktop.FromID(appID)
		if err != nil {
			logger.Log("error getting desktop entry %s: %v\n", window.AppID, err)
		}

		var icon string
		if desktopEntry != nil {
			icon = desktopEntry.Icon
		} else {
			icon = "application-x-executable"
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

	windowInt64, err := strconv.ParseInt(windowId, 10, 64)
	if err != nil {
		return fmt.Errorf("error focusing window: error reading window ID '%s' as integer: %w", windowId, err)
	}

	windowInt := int(windowInt64)

	if err := niri.FocusWindow(windowInt); err != nil {
		return fmt.Errorf("error switching to window %d: %w", windowInt, err)
	}

	return nil
}

func (w *WindowList) Prefix() string {
	return windowListPrefix
}
