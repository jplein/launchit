package state

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jplein/launchit/pkg/common/state/locations"
)

const (
	// Maximum number of recent entries to maintain
	maxRecent = 100
)

func Get() ([]string, error) {
	file, err := locations.RecentFilename()
	if err != nil {
		return nil, fmt.Errorf("error getting recent commands: %w", err)
	}

	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error getting recent commands: error reading from %s: %w", file, err)
	}

	recent := make([]string, 0)
	err = json.Unmarshal(contents, &recent)
	if err != nil {
		return nil, fmt.Errorf("error getting recent commands: error parsing %s as JSON: %w", file, err)
	}

	return recent, nil
}

func Add(id string) error {
	file, err := locations.RecentFilename()
	if err != nil {
		return fmt.Errorf("error adding '%s' to recent commands: %w", id, err)
	}

	recent, err := Get()
	if err != nil {
		recent = []string{}
	}

	updated := make([]string, 0)
	updated = append(updated, id)

	for _, recentCommand := range recent {
		if recentCommand != id {
			updated = append(updated, recentCommand)
		}
	}

	if len(updated) > maxRecent {
		updated = updated[:maxRecent]
	}

	output, err := json.Marshal(updated)
	if err != nil {
		return fmt.Errorf("error writing recent commands: error marshaling list to JSON: %w", err)
	}

	if err = os.WriteFile(file, output, 0o644); err != nil {
		return fmt.Errorf("error writing recent commands: %w", err)
	}

	return nil
}
