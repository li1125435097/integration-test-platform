package interpreters

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"integration-test-platform/server/src/store"
)

var ErrNotFound = errors.New("interpreter not found")

// Service manages interpreter configuration.
type Service struct {
	store *store.Store[File]
	path  string
}

// NewService creates a service backed by interpreters.json under dataDir.
func NewService(dataDir string) *Service {
	p := filepath.Join(dataDir, "interpreters.json")
	return &Service{
		store: store.NewJSON[File](p),
		path:  p,
	}
}

// EnsureDataDir creates empty interpreters.json if missing.
func (s *Service) EnsureDataDir() error {
	if _, err := os.Stat(s.path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return s.store.Save(File{Interpreters: []Interpreter{}})
	}
	return nil
}

// List returns all interpreters.
func (s *Service) List() ([]Interpreter, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	out := make([]Interpreter, len(f.Interpreters))
	copy(out, f.Interpreters)
	for i := range out {
		out[i].DefaultArgs = normalizeDefaultArgs(out[i].DefaultArgs)
	}
	return out, nil
}

type SaveInput struct {
	Language    string
	Path        string
	DefaultArgs []string
}

// Create adds a new interpreter.
func (s *Service) Create(in SaveInput) (*Interpreter, error) {
	if err := validateSaveInput(in); err != nil {
		return nil, err
	}
	abs, ok := absPath(strings.TrimSpace(in.Path))
	if !ok {
		return nil, fmt.Errorf("invalid interpreter path %q", in.Path)
	}
	now := time.Now().UTC()
	item := Interpreter{
		ID:          uuid.NewString(),
		Language:    in.Language,
		Path:        abs,
		DefaultArgs: normalizeDefaultArgs(in.DefaultArgs),
		Version:     resolveVersion(in.Language, abs, ""),
		UpdatedAt:   now,
	}
	if err := s.store.Update(func(f *File) error {
		f.Interpreters = append(f.Interpreters, item)
		return nil
	}); err != nil {
		return nil, err
	}
	return &item, nil
}

// Update modifies an interpreter by id.
func (s *Service) Update(id string, in SaveInput) (*Interpreter, error) {
	if err := validateSaveInput(in); err != nil {
		return nil, err
	}
	abs, ok := absPath(strings.TrimSpace(in.Path))
	if !ok {
		return nil, fmt.Errorf("invalid interpreter path %q", in.Path)
	}
	var updated Interpreter
	err := s.store.Update(func(f *File) error {
		idx := findIndex(f.Interpreters, id)
		if idx < 0 {
			return ErrNotFound
		}
		sc := &f.Interpreters[idx]
		prevPath := sc.Path
		sc.Language = in.Language
		sc.Path = abs
		sc.DefaultArgs = normalizeDefaultArgs(in.DefaultArgs)
		if prevPath != abs {
			sc.Version = resolveVersion(in.Language, abs, "")
		}
		if sc.IsDefault {
			clearDefaultExcept(f, sc.Language, id)
		}
		sc.UpdatedAt = time.Now().UTC()
		updated = *sc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// UpdatePath sets path only (preserves language and defaultArgs).
func (s *Service) UpdatePath(id, path, versionHint string) (*Interpreter, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}
	abs, ok := absPath(path)
	if !ok {
		return nil, fmt.Errorf("invalid interpreter path %q", path)
	}
	var updated Interpreter
	err := s.store.Update(func(f *File) error {
		idx := findIndex(f.Interpreters, id)
		if idx < 0 {
			return ErrNotFound
		}
		sc := &f.Interpreters[idx]
		sc.Path = abs
		sc.Version = resolveVersion(sc.Language, abs, versionHint)
		sc.UpdatedAt = time.Now().UTC()
		updated = *sc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// SetDefault toggles whether id is the default interpreter for its language.
func (s *Service) SetDefault(id string, isDefault bool) (*Interpreter, error) {
	var updated Interpreter
	err := s.store.Update(func(f *File) error {
		idx := findIndex(f.Interpreters, id)
		if idx < 0 {
			return ErrNotFound
		}
		sc := &f.Interpreters[idx]
		if isDefault {
			clearDefaultExcept(f, sc.Language, id)
			sc.IsDefault = true
		} else {
			sc.IsDefault = false
		}
		sc.UpdatedAt = time.Now().UTC()
		updated = *sc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// Delete removes an interpreter by id.
func (s *Service) Delete(id string) error {
	return s.store.Update(func(f *File) error {
		idx := findIndex(f.Interpreters, id)
		if idx < 0 {
			return ErrNotFound
		}
		f.Interpreters = append(f.Interpreters[:idx], f.Interpreters[idx+1:]...)
		return nil
	})
}

type BatchItem struct {
	Language    string
	Path        string
	DefaultArgs []string
	Version     string
}

type BatchResult struct {
	Created []Interpreter `json:"created"`
	Skipped int           `json:"skipped"`
}

// BatchCreate inserts items whose path is not already stored.
func (s *Service) BatchCreate(items []BatchItem) (*BatchResult, error) {
	if len(items) == 0 {
		return &BatchResult{Created: []Interpreter{}, Skipped: 0}, nil
	}
	result := &BatchResult{Created: []Interpreter{}}
	err := s.store.Update(func(f *File) error {
		existing := make(map[string]struct{}, len(f.Interpreters))
		for _, it := range f.Interpreters {
			existing[strings.ToLower(it.Path)] = struct{}{}
		}
		for _, raw := range items {
			in := SaveInput{
				Language:    raw.Language,
				Path:        raw.Path,
				DefaultArgs: raw.DefaultArgs,
			}
			if err := validateSaveInput(in); err != nil {
				return err
			}
			abs, ok := absPath(strings.TrimSpace(in.Path))
			if !ok {
				return fmt.Errorf("invalid interpreter path %q", in.Path)
			}
			key := strings.ToLower(abs)
			if _, dup := existing[key]; dup {
				result.Skipped++
				continue
			}
			now := time.Now().UTC()
			item := Interpreter{
				ID:          uuid.NewString(),
				Language:    in.Language,
				Path:        abs,
				DefaultArgs: normalizeDefaultArgs(in.DefaultArgs),
				Version:     resolveVersion(in.Language, abs, raw.Version),
				UpdatedAt:   now,
			}
			f.Interpreters = append(f.Interpreters, item)
			existing[key] = struct{}{}
			result.Created = append(result.Created, item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if result.Created == nil {
		result.Created = []Interpreter{}
	}
	return result, nil
}

func validateSaveInput(in SaveInput) error {
	if err := validateLanguage(in.Language); err != nil {
		return err
	}
	if strings.TrimSpace(in.Path) == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

func normalizeDefaultArgs(args []string) []string {
	if args == nil {
		return []string{}
	}
	out := make([]string, 0, len(args))
	for _, a := range args {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		out = append(out, a)
	}
	return out
}

func resolveVersion(language, path, stored string) string {
	stored = strings.TrimSpace(stored)
	if stored != "" {
		return stored
	}
	return DetectVersion(language, path)
}

func findIndex(list []Interpreter, id string) int {
	for i := range list {
		if list[i].ID == id {
			return i
		}
	}
	return -1
}

func clearDefaultExcept(f *File, language, keepID string) {
	for i := range f.Interpreters {
		if f.Interpreters[i].Language != language {
			continue
		}
		if f.Interpreters[i].ID == keepID {
			continue
		}
		f.Interpreters[i].IsDefault = false
	}
}
