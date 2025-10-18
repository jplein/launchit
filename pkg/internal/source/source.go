package source

type Entry struct {
	Description string
	ID          string
}

type Source interface {
	List() ([]Entry, error)
	Name() string
	Act(Entry) error
}
