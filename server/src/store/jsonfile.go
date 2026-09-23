package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Store persists a JSON document at a fixed path with atomic writes.
type Store[T any] struct {
	path string
	mu   sync.RWMutex
}

// NewJSON returns a store for the file at path.
func NewJSON[T any](path string) *Store[T] {
	return &Store[T]{path: path}
}

// Path returns the backing file path.
func (s *Store[T]) Path() string {
	return s.path
}

// Load reads the document. Missing file yields zero value and nil error.
func (s *Store[T]) Load() (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadUnlocked()
}

func (s *Store[T]) loadUnlocked() (T, error) {
	var zero T
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return zero, nil
		}
		return zero, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return zero, err
	}
	return v, nil
}

// Save writes the document atomically (temp file + rename).
func (s *Store[T]) Save(v T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveUnlocked(v)
}

func (s *Store[T]) saveUnlocked(v T) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Update loads, applies fn, and saves. fn runs under an exclusive lock.
func (s *Store[T]) Update(fn func(*T) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, err := s.loadUnlocked()
	if err != nil {
		return err
	}
	if err := fn(&v); err != nil {
		return err
	}
	return s.saveUnlocked(v)
}
