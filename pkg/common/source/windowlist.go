package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sort"
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

func historyFromServer() ([]uint64, error) {
	url := fmt.Sprintf("http://127.0.0.1:%s/api/v1/history", server.Port)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error reading from /api/v1/history: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error reading from /api/v1/history: invalid status code %d, response body: %s", resp.StatusCode, string(body))
	}

	history := []uint64{}
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("error parsing history JSON: %w", err)
	}

	return history, nil
}

// Sort a list of window descriptions in place, with the most recent windows appearing first
//
// windows: A list of windows
// history: A list of window IDs, ordered so that the most recent windows are at the end
func sortWindowsByHistory(windows []niri.WindowDescription, history []uint64) {
	// Create a map from window ID to its position in history
	historyPos := make(map[uint64]int)
	for i, id := range history {
		historyPos[id] = i
	}

	// Sort windows: most recently focused first
	sort.SliceStable(windows, func(i, j int) bool {
		idI := uint64(windows[i].ID)
		idJ := uint64(windows[j].ID)

		posI, inHistoryI := historyPos[idI]
		posJ, inHistoryJ := historyPos[idJ]

		// If both are in history, sort by position (higher = more recent = comes first)
		if inHistoryI && inHistoryJ {
			return posI > posJ
		}

		// If only one is in history, it comes first
		if inHistoryI {
			return true
		}
		if inHistoryJ {
			return false
		}

		// If neither is in history, maintain original order (stable sort handles this)
		return false
	})

}

func (w *WindowList) List() ([]Entry, error) {
	history := []uint64{}

	serverHistory, err := historyFromServer()
	if err != nil {
		logger.Log("error getting history from server: %v", err)
	} else {
		history = serverHistory
	}

	windows, err := niri.ListWindows()
	if err != nil {
		return nil, fmt.Errorf("error getting window list from Niri: %w", err)
	}

	sortWindowsByHistory(windows, history)

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
