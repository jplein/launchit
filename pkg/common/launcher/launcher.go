package launcher

import (
	"fmt"
	"io"
	"strings"

	"github.com/jplein/launchit/pkg/common/source"
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
		str += "\000" + "icon" + "\x1f" + icon
	}

	str = strings.ReplaceAll(str, "\n", " ")

	return str
}
