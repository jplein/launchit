package overrides

import (
	_ "embed"
	"fmt"

	"go.yaml.in/yaml/v4"
)

type Override struct {
	Icon     string `yaml:"icon"`
	AppID    string `yaml:"app-id"`
	WindowID string `yaml:"window-id"`
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

func ByIcon(icon string) (*Override, error) {
	overrides, err := getOverrides()
	if err != nil {
		return nil, err
	}

	for _, o := range overrides {
		if o.Icon == icon {
			return &o, nil
		}
	}

	return nil, nil
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

func ByWindowID(windowID string) (*Override, error) {
	overrides, err := getOverrides()
	if err != nil {
		return nil, err
	}

	for _, o := range overrides {
		if o.WindowID == windowID {
			return &o, nil
		}
	}

	return nil, nil
}
