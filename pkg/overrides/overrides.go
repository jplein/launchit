package overrides

import (
	_ "embed"
	"fmt"

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

func getOverrides() ([]Override, error) {
	var doc overridesDoc

	err := yaml.Unmarshal(overridesBuf, &doc)
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
