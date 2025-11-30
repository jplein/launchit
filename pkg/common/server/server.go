package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/jplein/launchit/pkg/common/logger"
)

const Port = "17324"

const (
	// Maximum number of requests allowed per minute for each handler
	maxRequestsPerMinute = 60
	// Time window for rate limiting
	rateLimitWindow = time.Minute
)

// LoadShedder implements rate limiting using a ring buffer of timestamps
type LoadShedder struct {
	timestamps []time.Time
	cursor     int
	mu         sync.Mutex
}

// NewLoadShedder creates a new LoadShedder with the specified capacity
func NewLoadShedder(capacity int) *LoadShedder {
	return &LoadShedder{
		timestamps: make([]time.Time, capacity),
		cursor:     0,
	}
}

// Allow checks if a request should be allowed based on rate limits
// Returns true if the request should be served, false if it should be shed (429)
func (ls *LoadShedder) Allow() bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	now := time.Now()

	// Calculate the next position in the ring buffer
	nextCursor := (ls.cursor + 1) % len(ls.timestamps)

	// Check the timestamp at the next position
	// If it's within the rate limit window, we've exceeded our limit
	oldestAllowed := now.Add(-rateLimitWindow)
	if !ls.timestamps[nextCursor].IsZero() && ls.timestamps[nextCursor].After(oldestAllowed) {
		// Too many requests in the time window
		return false
	}

	// Record this request and move the cursor
	ls.timestamps[nextCursor] = now
	ls.cursor = nextCursor

	return true
}

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
	lastEvent     string
	windowHistory []uint64
	mu            sync.RWMutex
}

func (n *NiriEventListener) Listen() error {
	go func() {
		const (
			initialBackoff = time.Second
			maxBackoff     = 64 * time.Second
			cooldownPeriod = 5 * time.Minute // Reset backoff if process runs this long
		)

		backoff := initialBackoff

		for {
			startTime := time.Now()

			cmd := exec.Command("niri", "msg", "-j", "event-stream")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				logger.Log("error creating stdout pipe for niri event-stream: %v, retrying in %v\n", err, backoff)
				time.Sleep(backoff)
				backoff = min(backoff*2, maxBackoff)
				continue
			}

			if err := cmd.Start(); err != nil {
				logger.Log("error starting niri event-stream: %v, retrying in %v\n", err, backoff)
				time.Sleep(backoff)
				backoff = min(backoff*2, maxBackoff)
				continue
			}

			logger.Log("niri event-stream process started\n")

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

			// Check how long the process ran
			uptime := time.Since(startTime)
			if uptime >= cooldownPeriod {
				// Process ran successfully for the cooldown period, reset backoff
				backoff = initialBackoff
				logger.Log("niri event-stream process exited after %v, resetting backoff, restarting in %v\n", uptime.Round(time.Second), backoff)
			} else {
				// Process failed quickly, use exponential backoff
				logger.Log("niri event-stream process exited after %v, restarting in %v\n", uptime.Round(time.Second), backoff)
				backoff = min(backoff*2, maxBackoff)
			}

			time.Sleep(backoff)
		}
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
var healthLoadShedder *LoadShedder
var historyLoadShedder *LoadShedder

func Start() error {
	eventListener = &NiriEventListener{}
	err := eventListener.Listen()
	if err != nil {
		return err
	}

	// Initialize load shedders for each handler
	healthLoadShedder = NewLoadShedder(maxRequestsPerMinute)
	historyLoadShedder = NewLoadShedder(maxRequestsPerMinute)

	http.HandleFunc("/api/v1/health", healthHandler)
	http.HandleFunc("/api/v1/history", historyHandler)

	addr := ":" + Port
	logger.Log("Starting server on %s\n", addr)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		return err
	}

	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if !healthLoadShedder.Allow() {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	if !historyLoadShedder.Allow() {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	}

	history := eventListener.WindowHistory()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}
