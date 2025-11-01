package source

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jplein/launchit/pkg/common/logger"
)

type Entry struct {
	Description string
	ID          string
	Icon        string
	Type        string
}

// Read an entry from a string. The string should contain a line with fields
// separated by tabs. The first field is th edescription, and the second is the
// ID. If the string contains more than one line, subsequent lines are ignored.
// Returns an error if the first line does not contain a tab, or if either of
// the two fields is empty.
func EntryFromString(s string) (Entry, error) {
	if len(s) == 0 {
		return Entry{}, fmt.Errorf("error reading entry from string: string is empty")
	}

	lines := strings.Split(s, "\n")

	firstLine := lines[0]
	if len(firstLine) == 0 {
		return Entry{}, fmt.Errorf("error reading entry from string: first line is empty")
	}

	fields := strings.Split(firstLine, "\t")
	if len(fields) == 1 {
		return Entry{}, fmt.Errorf("error reading entry from string: line does not contain a tab delimiter")
	}

	if len(fields) > 2 {
		return Entry{}, fmt.Errorf("error reading entry from string: line contains more than one tab-delimted field")
	}

	return Entry{Description: fields[0], ID: fields[1]}, nil
}

type Source interface {
	List() ([]Entry, error)
	Name() string
	Handle(entry Entry) error
	Prefix() string
}

type SourceSet struct {
	Sources []Source
}

func NewSourceSet(sources []Source) (*SourceSet, error) {
	if sources == nil {
		return nil, errors.New("invalid source list: nil")
	}

	if len(sources) == 0 {
		return nil, errors.New("invalid source list: source list is empty, expected at least one source")
	}

	// TODO: additional validation, make sure prefixes are unique?

	return &SourceSet{Sources: sources}, nil
}

func (s *SourceSet) List() ([]Entry, error) {
	entries := make([]Entry, 0)

	for _, src := range s.Sources {
		sourceEntries, err := src.List()
		if err != nil {
			logger.Log("%s\n", err.Error())
			continue
		}

		entries = append(entries, sourceEntries...)
	}

	return entries, nil
}

func DefaultSourceSet() (*SourceSet, error) {
	appSource := &Applications{}
	windowsSource := &WindowList{}
	commandsSource := &Commands{}

	return NewSourceSet([]Source{appSource, windowsSource, commandsSource})
}

func (s *SourceSet) Handle(entry Entry) error {
	id := entry.ID
	for _, source := range s.Sources {
		if strings.HasPrefix(id, source.Prefix()) {
			return source.Handle(entry)
		}
	}

	return fmt.Errorf("no handler found for %s", id)
}
