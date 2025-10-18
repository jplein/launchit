package source

import (
	"errors"
	"fmt"
	"os"
)

type Entry struct {
	Description string
	ID          string
}

type Source interface {
	List() ([]Entry, error)
	Name() string
	Act(Entry) error
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
			fmt.Fprint(os.Stderr, err.Error())
			continue
		}

		entries = append(entries, sourceEntries...)
	}

	return entries, nil
}

func DefaultSourceSet() (*SourceSet, error) {
	appSource := &Applications{}

	return NewSourceSet([]Source{appSource})
}
