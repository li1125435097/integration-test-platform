package kernelplans

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"integration-test-platform/server/src/store"
)

var (
	// ErrNotFound means the plan id is missing.
	ErrNotFound = errors.New("方案不存在")
	// ErrEmptyName means the plan name is blank after trim.
	ErrEmptyName = errors.New("方案名称不能为空")
	// ErrNameExists means another plan already uses the name.
	ErrNameExists = errors.New("方案名称已存在")
)

// Service manages persisted kernel-test plans.
type Service struct {
	store *store.Store[File]
	path  string
}

// NewService creates a service backed by kernel-test-plans.json under dataDir.
func NewService(dataDir string) *Service {
	p := filepath.Join(dataDir, "kernel-test-plans.json")
	return &Service{
		store: store.NewJSON[File](p),
		path:  p,
	}
}

// EnsureDataDir creates empty kernel-test-plans.json if missing.
func (s *Service) EnsureDataDir() error {
	if _, err := os.Stat(s.path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return s.store.Save(File{Plans: []Plan{}})
	}
	return nil
}

// List returns all plans, most recently updated first.
func (s *Service) List() ([]Plan, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	out := make([]Plan, len(f.Plans))
	copy(out, f.Plans)
	for i := range out {
		out[i].Form = normalizeForm(out[i].Form)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

// Get returns one plan by id.
func (s *Service) Get(id string) (*Plan, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	for i := range f.Plans {
		if f.Plans[i].ID == id {
			plan := f.Plans[i]
			plan.Form = normalizeForm(plan.Form)
			return &plan, nil
		}
	}
	return nil, ErrNotFound
}

// Create stores a new plan. Name is trimmed and must be unique.
func (s *Service) Create(name string, form Form) (*Plan, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyName
	}
	now := time.Now().UTC()
	plan := Plan{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Form:      normalizeForm(form),
	}
	err := s.store.Update(func(f *File) error {
		if f.Plans == nil {
			f.Plans = []Plan{}
		}
		for _, item := range f.Plans {
			if item.Name == name {
				return ErrNameExists
			}
		}
		f.Plans = append(f.Plans, plan)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

// Update replaces the name and form of an existing plan.
func (s *Service) Update(id, name string, form Form) (*Plan, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyName
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	var saved Plan
	err := s.store.Update(func(f *File) error {
		idx := -1
		for i := range f.Plans {
			if f.Plans[i].ID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			return ErrNotFound
		}
		for i := range f.Plans {
			if i != idx && f.Plans[i].Name == name {
				return ErrNameExists
			}
		}
		f.Plans[idx].Name = name
		f.Plans[idx].Form = normalizeForm(form)
		f.Plans[idx].UpdatedAt = time.Now().UTC()
		saved = f.Plans[idx]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

// Delete removes a plan by id.
func (s *Service) Delete(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	return s.store.Update(func(f *File) error {
		next := make([]Plan, 0, len(f.Plans))
		found := false
		for _, item := range f.Plans {
			if item.ID == id {
				found = true
				continue
			}
			next = append(next, item)
		}
		if !found {
			return ErrNotFound
		}
		f.Plans = next
		return nil
	})
}

func normalizeForm(form Form) Form {
	form.Sources = nonNil(form.Sources)
	form.Kernels = nonNil(form.Kernels)
	form.Flags = nonNil(form.Flags)
	form.CustomFlags = nonNil(form.CustomFlags)
	return form
}

func nonNil(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}
