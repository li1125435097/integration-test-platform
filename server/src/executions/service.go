package executions

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"

	"integration-test-platform/server/src/store"
)

var ErrNotFound = errors.New("execution record not found")

// Service manages persisted script execution records.
type Service struct {
	store *store.Store[File]
	path  string
}

// NewService creates a service backed by execution-records.json under dataDir.
func NewService(dataDir string) *Service {
	p := filepath.Join(dataDir, "execution-records.json")
	return &Service{
		store: store.NewJSON[File](p),
		path:  p,
	}
}

// EnsureDataDir creates empty execution-records.json if missing.
func (s *Service) EnsureDataDir() error {
	if _, err := os.Stat(s.path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return s.store.Save(File{Records: []Record{}})
	}
	return nil
}

// List returns all records, newest first.
func (s *Service) List() ([]Record, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	out := make([]Record, len(f.Records))
	copy(out, f.Records)
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

// Get returns one record by id.
func (s *Service) Get(id string) (*Record, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	for i := range f.Records {
		if f.Records[i].ID == id {
			rec := f.Records[i]
			return &rec, nil
		}
	}
	return nil, ErrNotFound
}

// Add appends a record and returns it with generated id.
func (s *Service) Add(rec Record) (*Record, error) {
	if rec.ID == "" {
		rec.ID = uuid.NewString()
	}
	var saved Record
	err := s.store.Update(func(f *File) error {
		f.Records = append(f.Records, rec)
		saved = rec
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}
