package launcher

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/jplein/launchit/pkg/common/logger"
	"github.com/jplein/launchit/pkg/common/source"
	"github.com/jplein/launchit/pkg/common/state"
)

type Launcher struct {
	sources source.SourceSet
}

func NewLauncher(sources source.SourceSet) (Launcher, error) {
	return Launcher{
		sources: sources,
	}, nil
}

func (l *Launcher) List() ([]source.Entry, error) {
	entries, err := l.sources.List()
	if err != nil {
		return nil, fmt.Errorf("error listing entries: %w", err)
	}

	recents, err := state.Get()
	if err != nil {
		logger.Log("error reading recents: %v", err)
	}

	if recents != nil {
		slices.SortFunc(entries, func(a, b source.Entry) int {
			indexA := slices.Index(recents, a.ID)
			indexB := slices.Index(recents, b.ID)

			// Both found in recents: sort by index (lower index comes first)
			if indexA != -1 && indexB != -1 {
				return indexA - indexB
			}

			// Only a is in recents: a comes first
			if indexA != -1 {
				return -1
			}

			// Only b is in recents: b comes first
			if indexB != -1 {
				return 1
			}

			// Neither in recents: maintain original order
			return 0
		})
	}

	return entries, nil
}

func (l *Launcher) Write(writer io.Writer) error {
	entries, err := l.List()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		_, err := writer.Write([]byte(getLine(entry) + "\n"))
		if err != nil {
			return fmt.Errorf("error writing entries: %w", err)
		}
	}

	return nil
}

func getLine(entry source.Entry) string {
	str := fmt.Sprintf(
		"%s\t%s",
		strings.ReplaceAll(entry.Description, "\t", "    "),
		strings.ReplaceAll(entry.ID, "\t", "    "),
	)

	if entry.Icon != "" {
		icon := strings.ReplaceAll(entry.Icon, "\t", "")
		str += "\x00" + "icon" + "\x1f" + icon
	}

	str = strings.ReplaceAll(str, "\n", " ")

	return str
}
