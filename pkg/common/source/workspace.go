package source

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/source/niri"
)

const (
	workspaceSourceName   = "workspaces"
	workspaceSourceType   = "workspace"
	workspacePrefix       = "workspace"
	workspaceSwitchSuffix = "switch"
	workspaceMoveSuffix   = "move"
)

type Workspaces struct{}

func (w *Workspaces) List() ([]Entry, error) {
	workspaces, err := niri.ListWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("error getting workspace list from Niri: %w", err)
	}

	entries := make([]Entry, 0)

	for _, workspace := range workspaces {
		switchEntry := Entry{
			Description: fmt.Sprintf("Niri: Switch to workspace %d", workspace.Index),
			ID:          fmt.Sprintf("%s:%d%s", workspacePrefix, workspace.ID, workspaceSwitchSuffix),
			Type:        workspaceSourceType,
		}

		entries = append(entries, switchEntry)

		moveEntry := Entry{
			Description: fmt.Sprintf("Niri: Move active window to workspace %d", workspace.Index),
			ID:          fmt.Sprintf("%s:%d%s", workspacePrefix, workspace.ID, workspaceMoveSuffix),
			Type:        workspaceSourceType,
		}

		entries = append(entries, moveEntry)
	}

	return entries, nil
}

func (w *Workspaces) Name() string {
	return workspaceSourceName
}

func (w *Workspaces) Handle(entry Entry) error {
	id := entry.ID

	if !strings.HasPrefix(id, workspacePrefix) {
		return fmt.Errorf("not a Niri workspace: %s", id)
	}

	widAndSuffix := id[len(workspacePrefix)+1:]
	if widAndSuffix == "" {
		return fmt.Errorf("not a valid ID: No workspace ID and type suffix")
	}

	var cmd *exec.Cmd
	switch {
	case strings.HasSuffix(widAndSuffix, workspaceSwitchSuffix):
		wid := widAndSuffix[:(len(widAndSuffix) - len(workspaceSwitchSuffix))]
		cmd = exec.Command("niri", "msg", "action", "focus-workspace", wid)
	case strings.HasSuffix(widAndSuffix, workspaceMoveSuffix):
		wid := widAndSuffix[:(len(widAndSuffix) - len(workspaceMoveSuffix))]
		cmd = exec.Command("niri", "msg", "action", "move-window-to-workspace", wid)
	default:
		return fmt.Errorf("not a valid ID: does not end with %s or %s", workspaceSwitchSuffix, workspaceMoveSuffix)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stdout.Len() > 0 {
			logger.Log("niri focus-window stdout: %s\n", stdout.String())
		}
		if stderr.Len() > 0 {
			logger.Log("niri focus-window stderr: %s\n", stderr.String())
		}
		return fmt.Errorf("error running command %v", cmd)
	}

	return nil
}

func (w *Workspaces) Prefix() string {
	return workspacePrefix
}
