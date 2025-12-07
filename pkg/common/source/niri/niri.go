package niri

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/jplein/launchit/pkg/common/logger"
)

type WindowDescription struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	AppID string `json:"app_id"`
}

func ListWindows() ([]WindowDescription, error) {
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

	return windows, nil
}

type WorkspaceDescription struct {
	ID        int     `json:"id"`
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
