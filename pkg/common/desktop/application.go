package desktop

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/ini.v1"
)

const (
	defaultXDGDataDirs = "/usr/local/share:/usr/share"
	defaultXDGDataHome = "~/.local/share"
)

type App struct {
	Icon     string
	Name     string
	ID       string
	Filename string
}

func List() ([]App, error) {
	dirs, err := getSearchDirs()
	if err != nil {
		return nil, err
	}

	apps := make([]App, 0)

	for _, dir := range dirs {
		files, err := getDesktopFiles(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", dir, err)
			continue
		}

		for _, file := range files {
			app, err := getEntry(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading desktop file %s: %v\n", file, err)
				continue
			}

			apps = append(apps, app)
		}
	}

	return apps, nil
}

func Get(id string) (*App, error) {
	apps, err := List()
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.ID == id {
			return &app, nil
		}
	}

	return nil, errors.New("no application with ID '" + id + "' found")
}

func getEntry(desktopFile string) (App, error) {
	desktopFileEntry, err := ini.Load(desktopFile)
	if err != nil {
		return App{}, fmt.Errorf("error reading from %s: %w", desktopFile, err)
	}

	desktopSection := desktopFileEntry.Section("Desktop Entry")

	name := desktopSection.Key("Name").String()
	if name == "" {
		return App{}, fmt.Errorf("error reading from %s: no Name found in [Desktop Entry] section", desktopFile)
	}

	icon := desktopSection.Key("Icon").String()
	if icon == "" {
		icon = "application-x-executable"
	}

	basename := strings.TrimSuffix(path.Base(desktopFile), ".desktop")

	return App{
		Icon:     icon,
		Name:     name,
		Filename: desktopFile,
		ID:       basename,
	}, nil
}

func getSearchDirs() ([]string, error) {
	xdgDataDirs := getXDGDataDirs()
	xdgDataHome, err := getXDGDataHome()
	if err != nil {
		return nil, err
	}

	searchDirs := []string{path.Join(xdgDataHome, "applications")}

	xdgDataDirsEntries := strings.Split(xdgDataDirs, ":")
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
