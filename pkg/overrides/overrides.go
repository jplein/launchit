package overrides

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/jplein/launchit/pkg/common/state/locations"
	"go.yaml.in/yaml/v4"
)

type Override struct {
	// The basename of the .desktop file
	AppID string `yaml:"app-id"`

	// The application ID as returned by Niri or another window manager for windows of this application
	WindowAppID string `yaml:"window-app-id"`
}

type overridesDoc struct {
	Overrides []Override `yaml:"overrides"`
}

//go:embed res/overrides.yaml
var overridesBuf []byte

const (
	// Path to the overrides file, relative to the XDG config directory
	overridesFile = "overrides.yaml"
)

func getOverrides() ([]Override, error) {
	overridesPath, err := locations.Initialize(locations.XDGConfigDir, overridesFile, overridesBuf, locations.DefaultFilePermission)
	if err != nil {
		return nil, fmt.Errorf("error getting overrides: %w", err)
	}

	buf, err := os.ReadFile(overridesPath)
	if err != nil {
		return nil, fmt.Errorf("error getting overrides: error reading from %s: %w", overridesPath, err)
	}

	var doc overridesDoc

	err = yaml.Unmarshal(buf, &doc)
	if err != nil {
		return nil, fmt.Errorf("error reading overrides: error parsing YAML: %w", err)
	}

	return doc.Overrides, nil
}

func ByAppID(appID string) (*Override, error) {
	overrides, err := getOverrides()
	if err != nil {
		return nil, err
	}

	for _, o := range overrides {
		if o.AppID == appID {
			return &o, nil
		}
	}

	return nil, nil
}

func ByWindowAppID(windowID string) (*Override, error) {
	overrides, err := getOverrides()
	if err != nil {
		return nil, err
	}

	for _, o := range overrides {
		if o.WindowAppID == windowID {
			return &o, nil
		}
	}

	return nil, nil
}
