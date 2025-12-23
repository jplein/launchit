package source

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/state/locations"
	"go.yaml.in/yaml/v4"
)

const (
	commandsSourceName = "commands"
	commandsSourceType = "Command"
	commandPrefix      = "command"
	commandDefaultIcon = "applications-system-symbolic"
)

type command struct {
	Executable  string   `yaml:"executable"`
	Args        []string `yaml:"args"`
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Icon        string   `yaml:"icon"`
}

type Commands struct {
	commands []command
}

func (c *Commands) List() ([]Entry, error) {
	if err := c.readCommands(); err != nil {
		return nil, err
	}

	entries := make([]Entry, 0)
	for _, command := range c.commands {
		entry := Entry{
			Description: command.Description,
			ID:          commandPrefix + ":" + command.ID,
			Icon:        command.Icon,
			Type:        commandsSourceType,
		}
		if entry.Icon == "" {
			entry.Icon = commandDefaultIcon
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (c *Commands) Handle(entry Entry) error {
	if err := c.readCommands(); err != nil {
		return err
	}

	for _, c := range c.commands {
		if commandPrefix+":"+c.ID == entry.ID {
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(c.Executable, c.Args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				logger.Log("command stdout: %s\n", stdout.String())
				logger.Log("command stderr: %s\n", stderr.String())
				return fmt.Errorf("error running command: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("error running command: no command found with id %s", entry.ID)
}

func (c *Commands) Name() string {
	return commandsSourceName
}

func (c *Commands) Prefix() string {
	return commandPrefix
}

const (
	configFile = "commands.yaml" // Relative to the config directory
)

func (c *Commands) commandsFile() (string, error) {
	configDir, err := locations.ConfigDirectory()
	if err != nil {
		return "", fmt.Errorf("error getting commands: error getting config directory: %w", err)
	}

	file := path.Join(configDir, configFile)

	return file, nil
}

func (c *Commands) readCommands() error {
	if c.commands != nil {
		return nil
	}

	file, err := c.commandsFile()
	if err != nil {
		return err
	}

	buf, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading commands: %w", err)
	}

	var yamlCommands []command
	if err := yaml.Unmarshal(buf, &yamlCommands); err != nil {
		return fmt.Errorf("error reading commands from YAML: %w", err)
	}

	c.commands = yamlCommands
	return nil
}
