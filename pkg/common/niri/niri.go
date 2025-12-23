package niri

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sort"

	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/server"
)

type WindowDescription struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	AppID string `json:"app_id"`
}

func ListWindows(sortWindows bool) ([]WindowDescription, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("niri", "msg", "--json", "windows")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stdout.Len() > 0 {
			logger.Log("niri windows stdout: %s\n", stdout.String())
		}
		if stderr.Len() > 0 {
			logger.Log("niri windows stderr: %s\n", stderr.String())
		}
		return nil, fmt.Errorf("error getting windows from Niri: %w", err)
	}

	listBytes := stdout.Bytes()

	windows := make([]WindowDescription, 0)
	if err := json.Unmarshal(listBytes, &windows); err != nil {
		logger.Log("niri window list JSON output:\n")
		logger.Log(string(listBytes))
		return nil, fmt.Errorf("error getting windows from Niri: error parsing JSON: %w", err)
	}

	if sortWindows {
		history := []uint64{}
		serverHistory, err := historyFromServer()
		if err != nil {
			logger.Log("error getting history from server: %v", err)
		} else {
			history = serverHistory
		}

		sortWindowsByHistory(windows, history)
	}

	return windows, nil
}

func FocusWindow(windowID int) error {
	windowStr := fmt.Sprintf("%d", windowID)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("niri", "msg", "action", "focus-window", "--id", windowStr)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stdout.Len() > 0 {
			logger.Log("niri focus-window stdout: %s\n", stdout.String())
		}
		if stderr.Len() > 0 {
			logger.Log("niri focus-window stderr: %s\n", stderr.String())
		}
		return fmt.Errorf("error switching to window %s: %w", windowID, err)
	}

	if stdout.Len() > 0 {
		logger.Log("niri focus-window stdout: %s\n", stdout.String())
	}
	if stderr.Len() > 0 {
		logger.Log("niri focus-window stderr: %s\n", stderr.String())
	}

	return nil
}

type WorkspaceDescription struct {
	ID        int     `json:"id"`
	Index     int     `json:"idx"`
	Name      *string `json:"name"`
	IsActive  bool    `json:"is_active"`
	IsFocused bool    `json:"is_focused"`
}

func ListWorkspaces() ([]WorkspaceDescription, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("niri", "msg", "--json", "workspaces")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stdout.Len() > 0 {
			logger.Log("niri workspaces stdout: %s\n", stdout.String())
		}
		if stderr.Len() > 0 {
			logger.Log("niri workspaces stdout: %s\n", stderr.String())
		}
		return nil, fmt.Errorf("error getting workspaces from Niri: %w", err)
	}

	listBytes := stdout.Bytes()

	workspaces := make([]WorkspaceDescription, 0)
	if err := json.Unmarshal(listBytes, &workspaces); err != nil {
		logger.Log("niri window list JSON output:\n")
		logger.Log(string(listBytes))
		return nil, fmt.Errorf("error getting windows from Niri: error parsing JSON: %w", err)
	}
	return workspaces, nil
}

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
func sortWindowsByHistory(windows []WindowDescription, history []uint64) {
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
