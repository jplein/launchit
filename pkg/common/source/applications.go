package source

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/ini.v1"
)

type Applications struct{}

const (
	defaultXDGDataDirs = "/usr/local/share:/usr/share"
	defaultXDGDataHome = "~/.local/share"
	idPrefix           = "app"
	sourceName         = "applications"
)

func (a *Applications) List() ([]Entry, error) {
	searchDirs, err := getSearchDirs()
	if err != nil {
		return nil, err
	}

	entries := getEntries(searchDirs)
	return entries, nil
}

func (a *Applications) Name() string {
	return sourceName
}

func (a *Applications) Handle(entry Entry) error {
	id := entry.ID
	fmt.Fprintf(os.Stderr, "handling application with id %s\n", id)

	if !strings.HasPrefix(id, idPrefix+":") {
		return fmt.Errorf("not an application: %s", id)
	}

	filename := id[len(idPrefix)+1:]
	if filename == "" {
		return fmt.Errorf("not a valid ID: filename is empty: %s", id)
	}

	cmd := exec.Command("gio", "launch", filename)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch application %s: %w", filename, err)
	}

	return nil
}

func (a *Applications) Prefix() string {
	return idPrefix
}

// Returns a list of directories in which to look for application .desktopp files
func getSearchDirs() ([]string, error) {
	xdgDataDirs := getXDGDataDirs()
	xdgDataHome, err := getXDGDataHome()
	if err != nil {
		return nil, err
	}

	xdgDataDirsEntries := strings.Split(xdgDataDirs, ":")
	searchDirs := []string{path.Join(xdgDataHome, "applications")}

	for _, entry := range xdgDataDirsEntries {
		searchDirs = append(searchDirs, path.Join(entry, "applications"))
	}

	return searchDirs, nil
}

func getXDGDataDirs() string {
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")
	if xdgDataDirs == "" {
		xdgDataDirs = defaultXDGDataDirs
	}

	return xdgDataDirs
}

func getXDGDataHome() (string, error) {
	var xdgDataHome string

	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("erorr getting home directory: HOME environment variable not set")
	}

	xdgDataHome = os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		xdgDataHome = defaultXDGDataHome
	}

	if strings.HasPrefix(xdgDataHome, "~") {
		xdgDataHome = home + xdgDataHome[1:]
	}

	return xdgDataHome, nil
}

func getEntries(searchDirs []string) []Entry {
	entries := make([]Entry, 0)

	foundMap := make(map[string]bool)

	for _, dir := range searchDirs {
		desktopFiles, err := getDesktopFiles(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		for _, desktopFile := range desktopFiles {
			entry, err := getEntry(desktopFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}

			if !foundMap[entry.Description] {
				entries = append(entries, entry)
				foundMap[entry.Description] = true
			}
		}
	}

	return entries
}

func getDesktopFiles(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", dir, err)
	}

	desktopFiles := make([]string, 0)
	for _, entry := range dirEntries {
		if strings.HasSuffix(entry.Name(), ".desktop") {
			desktopFiles = append(desktopFiles, path.Join(dir, entry.Name()))
		}
	}

	return desktopFiles, nil
}

func getEntry(desktopFile string) (Entry, error) {
	desktopFileEntry, err := ini.Load(desktopFile)
	if err != nil {
		return Entry{}, fmt.Errorf("error reading from %s: %w", desktopFile, err)
	}

	desktopSection := desktopFileEntry.Section("Desktop Entry")

	name := desktopSection.Key("Name").String()
	if name == "" {
		return Entry{}, fmt.Errorf("error reading from %s: no Name found in [Desktop Entry] section", desktopFile)
	}

	icon := desktopSection.Key("Icon").String()
	if icon == "" {
		icon = "application-x-executable"
	}

	return Entry{
		Description: name,
		ID:          idPrefix + ":" + desktopFile,
		Icon:        icon,
	}, nil
}
