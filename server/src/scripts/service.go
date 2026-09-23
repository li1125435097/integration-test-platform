package scripts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"integration-test-platform/server/src/store"
)

var (
	ErrNotFound        = errors.New("script not found")
	ErrVersionNotFound = errors.New("version not found")
)

// Service manages script metadata and files.
type Service struct {
	store      *store.Store[File]
	filesRoot  string
	scriptsPath string
}

// NewService creates a script service. scriptsJSON is the full path to scripts.json.
func NewService(dataDir string) *Service {
	scriptsPath := filepath.Join(dataDir, "scripts.json")
	return &Service{
		store:       store.NewJSON[File](scriptsPath),
		filesRoot:   filepath.Join(dataDir, "script-files"),
		scriptsPath: scriptsPath,
	}
}

// EnsureDataDir creates data directories and empty scripts.json if missing.
func (s *Service) EnsureDataDir() error {
	if err := os.MkdirAll(s.filesRoot, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(s.scriptsPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return s.store.Save(File{Scripts: []Script{}})
	}
	return nil
}

func (s *Service) scriptDir(id string) string {
	return filepath.Join(s.filesRoot, id)
}

func (s *Service) currentPath(id, fileName string) string {
	return filepath.Join(s.scriptDir(id), fileName)
}

func (s *Service) versionPath(id, fileName string) string {
	return filepath.Join(s.scriptDir(id), "versions", fileName)
}

// List returns all scripts without file content.
func (s *Service) List() ([]ListItem, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	out := make([]ListItem, 0, len(f.Scripts))
	for _, sc := range f.Scripts {
		out = append(out, s.listItemFor(sc))
	}
	return out, nil
}

func (sc Script) toListItem() ListItem {
	return ListItem{
		ID:             sc.ID,
		Name:           sc.Name,
		Description:    sc.Description,
		Language:       sc.Language,
		HasInterpreter: sc.HasInterpreter,
		UpdatedAt:      sc.UpdatedAt,
	}
}

func (s *Service) listItemFor(sc Script) ListItem {
	item := sc.toListItem()
	item.CurrentVersion = s.matchedVersion(sc)
	return item
}

// matchedVersion returns the newest snapshot whose file bytes equal the current script.
func (s *Service) matchedVersion(sc Script) string {
	current, err := os.ReadFile(s.currentPath(sc.ID, sc.FileName))
	if err != nil {
		return ""
	}
	for i := len(sc.Versions) - 1; i >= 0; i-- {
		v := sc.Versions[i]
		snap, err := os.ReadFile(s.versionPath(sc.ID, v.FileName))
		if err != nil {
			continue
		}
		if bytes.Equal(current, snap) {
			return v.ID
		}
	}
	return ""
}

// Get returns metadata and current file content.
func (s *Service) Get(id string) (*Detail, error) {
	sc, err := s.findScript(id)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(s.currentPath(id, sc.FileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			content = []byte{}
		} else {
			return nil, err
		}
	}
	d := &Detail{
		ListItem: s.listItemFor(*sc),
		Content:  string(content),
	}
	return d, nil
}

type CreateInput struct {
	Name        string
	Description string
	Language    string
	Content     string
}

// Create adds a new script with initial file content.
func (s *Service) Create(in CreateInput) (*Detail, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	ext, err := ExtForLanguage(in.Language)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	fileName := "current." + ext
	dir := s.scriptDir(id)
	if err := os.MkdirAll(filepath.Join(dir, "versions"), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(s.currentPath(id, fileName), []byte(in.Content), 0o644); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	sc := Script{
		ID:             id,
		Name:           in.Name,
		Description:    in.Description,
		Language:       in.Language,
		HasInterpreter: false,
		FileName:       fileName,
		UpdatedAt:      now,
		Versions:       []Version{},
	}
	if err := s.store.Update(func(f *File) error {
		f.Scripts = append(f.Scripts, sc)
		return nil
	}); err != nil {
		return nil, err
	}
	return &Detail{ListItem: s.listItemFor(sc), Content: in.Content}, nil
}

type UpdateInput struct {
	Name        string
	Description string
	Language    string
	Content     string
}

// Update saves metadata and current file; renames current file if language changes.
func (s *Service) Update(id string, in UpdateInput) (*Detail, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	ext, err := ExtForLanguage(in.Language)
	if err != nil {
		return nil, err
	}
	newFileName := "current." + ext

	var updated Script
	err = s.store.Update(func(f *File) error {
		idx := -1
		for i := range f.Scripts {
			if f.Scripts[i].ID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			return ErrNotFound
		}
		sc := &f.Scripts[idx]
		oldPath := s.currentPath(id, sc.FileName)
		sc.Name = in.Name
		sc.Description = in.Description
		sc.Language = in.Language
		sc.UpdatedAt = time.Now().UTC()
		if sc.FileName != newFileName {
			newPath := s.currentPath(id, newFileName)
			if err := moveFile(oldPath, newPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			sc.FileName = newFileName
		}
		if err := os.WriteFile(s.currentPath(id, sc.FileName), []byte(in.Content), 0o644); err != nil {
			return err
		}
		updated = *sc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Detail{ListItem: s.listItemFor(updated), Content: in.Content}, nil
}

// Delete removes a script from metadata and deletes its on-disk files.
func (s *Service) Delete(id string) error {
	if err := s.store.Update(func(f *File) error {
		for i := range f.Scripts {
			if f.Scripts[i].ID != id {
				continue
			}
			f.Scripts = append(f.Scripts[:i], f.Scripts[i+1:]...)
			return nil
		}
		return ErrNotFound
	}); err != nil {
		return err
	}
	if err := os.RemoveAll(s.scriptDir(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	return copyFile(src, dst)
}

func (s *Service) findScript(id string) (*Script, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	for i := range f.Scripts {
		if f.Scripts[i].ID == id {
			sc := f.Scripts[i]
			return &sc, nil
		}
	}
	return nil, ErrNotFound
}

// ListVersions returns version metadata for a script.
func (s *Service) ListVersions(id string) ([]Version, error) {
	sc, err := s.findScript(id)
	if err != nil {
		return nil, err
	}
	out := make([]Version, len(sc.Versions))
	copy(out, sc.Versions)
	return out, nil
}

// AddVersion copies the current script file into versions/.
func (s *Service) AddVersion(id string, name, remark string) (*Version, error) {
	if name == "" {
		name = time.Now().UTC().Format("20060102150405")
	}
	sc, err := s.findScript(id)
	if err != nil {
		return nil, err
	}
	ext, err := ExtForLanguage(sc.Language)
	if err != nil {
		return nil, err
	}
	versionFileName := name + "." + ext
	src := s.currentPath(id, sc.FileName)
	dst := s.versionPath(id, versionFileName)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return nil, err
	}
	if err := copyFile(src, dst); err != nil {
		return nil, err
	}
	ver := Version{
		ID:        name,
		FileName:  versionFileName,
		Remark:    remark,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.Update(func(f *File) error {
		for i := range f.Scripts {
			if f.Scripts[i].ID != id {
				continue
			}
			for _, existing := range f.Scripts[i].Versions {
				if existing.ID == name {
					return fmt.Errorf("version %q already exists", name)
				}
			}
			f.Scripts[i].Versions = append(f.Scripts[i].Versions, ver)
			return nil
		}
		return ErrNotFound
	}); err != nil {
		return nil, err
	}
	return &ver, nil
}

// Restore copies a version file over the current script file.
func (s *Service) Restore(id, versionID string) (*ListItem, error) {
	var item ListItem
	err := s.store.Update(func(f *File) error {
		idx := -1
		for i := range f.Scripts {
			if f.Scripts[i].ID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			return ErrNotFound
		}
		sc := &f.Scripts[idx]
		var ver *Version
		for i := range sc.Versions {
			if sc.Versions[i].ID == versionID {
				ver = &sc.Versions[i]
				break
			}
		}
		if ver == nil {
			return ErrVersionNotFound
		}
		src := s.versionPath(id, ver.FileName)
		dst := s.currentPath(id, sc.FileName)
		if err := copyFile(src, dst); err != nil {
			return err
		}
		sc.UpdatedAt = time.Now().UTC()
		item = s.listItemFor(*sc)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
