package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os/exec"
	"sync"

	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/source/niri"
)

const defaultPort = "17324"

type NiriEvent struct {
	WindowFocusChanged *struct {
		ID uint64 `json:"id"`
	} `json:"WindowFocusChanged"`
	WindowClosed *struct {
		ID uint64 `json:"id"`
	} `json:"WindowClosed"`
	WindowOpenedOrChanged *struct {
		Window struct {
			ID uint64 `json:"id"`
		} `json:"window"`
	} `json:"WindowOpenedOrChanged"`
}

type NiriEventListener struct {
	lastEvent      string
	windowHistory  []uint64
	mu             sync.RWMutex
}

func (n *NiriEventListener) Listen() error {
	cmd := exec.Command("niri", "msg", "-j", "event-stream")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			n.mu.Lock()
			n.lastEvent = line
			n.handleEvent(line)
			n.mu.Unlock()
		}

		if err := scanner.Err(); err != nil {
			logger.Log("error reading from niri event-stream: %v\n", err)
		}

		cmd.Wait()
	}()

	return nil
}

func (n *NiriEventListener) handleEvent(line string) {
	var event NiriEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return
	}

	if event.WindowFocusChanged != nil {
		windowID := event.WindowFocusChanged.ID
		n.addWindowToHistory(windowID)
	} else if event.WindowClosed != nil {
		windowID := event.WindowClosed.ID
		n.removeWindowFromHistory(windowID)
	} else if event.WindowOpenedOrChanged != nil {
		windowID := event.WindowOpenedOrChanged.Window.ID
		n.addWindowToHistory(windowID)
	}
}

func (n *NiriEventListener) addWindowToHistory(windowID uint64) {
	// Remove the window ID if it already exists
	for i, id := range n.windowHistory {
		if id == windowID {
			n.windowHistory = append(n.windowHistory[:i], n.windowHistory[i+1:]...)
			break
		}
	}

	// Add the window ID to the end of the list
	n.windowHistory = append(n.windowHistory, windowID)
}

func (n *NiriEventListener) removeWindowFromHistory(windowID uint64) {
	// Remove the window ID from the history
	for i, id := range n.windowHistory {
		if id == windowID {
			n.windowHistory = append(n.windowHistory[:i], n.windowHistory[i+1:]...)
			break
		}
	}
}

func (n *NiriEventListener) LastEvent() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.lastEvent
}

func (n *NiriEventListener) WindowHistory() []uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	// Return a copy to prevent external modification
	history := make([]uint64, len(n.windowHistory))
	copy(history, n.windowHistory)
	return history
}

var eventListener *NiriEventListener

func Start() error {
	eventListener = &NiriEventListener{}
	err := eventListener.Listen()
	if err != nil {
		return err
	}

	http.HandleFunc("/api/v1/health", healthHandler)
	http.HandleFunc("/api/v1/last-event", lastEventHandler)
	http.HandleFunc("/api/v1/history", historyHandler)
	http.HandleFunc("/api/v1/windows", windowsHandler)

	addr := ":" + defaultPort
	logger.Log("Starting server on %s\n", addr)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		return err
	}

	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func lastEventHandler(w http.ResponseWriter, r *http.Request) {
	lastEvent := eventListener.LastEvent()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(lastEvent))
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	history := eventListener.WindowHistory()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}

func windowsHandler(w http.ResponseWriter, r *http.Request) {
	windows, err := niri.ListWindows()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(windows)
}
