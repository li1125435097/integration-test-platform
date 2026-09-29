package scripts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	store       *store.Store[File]
	filesRoot   string
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
		ID:            sc.ID,
		Name:          sc.Name,
		Description:   sc.Description,
		Language:      sc.Language,
		InterpreterID: sc.InterpreterID,
		UpdatedAt:     sc.UpdatedAt,
		Files:         fileEntries(effectiveFiles(sc)),
		Variables:     sc.Variables,
	}
}

func (s *Service) listItemFor(sc Script) ListItem {
	item := sc.toListItem()
	item.CurrentVersion = s.matchedVersion(sc)
	return item
}

// matchedVersion returns the newest snapshot that matches the current workspace.
func (s *Service) matchedVersion(sc Script) string {
	locals, err := s.localBytes(sc)
	if err != nil {
		return ""
	}
	currentMeta := effectiveFiles(sc)
	hasExtraLocal := false
	for _, f := range currentMeta {
		if f.Kind == FileKindLocal {
			hasExtraLocal = true
			break
		}
	}
	mainName, _ := MainFileName(sc.Language)

	for i := len(sc.Versions) - 1; i >= 0; i-- {
		v := sc.Versions[i]
		if s.snapshotIsDir(sc.ID, v) {
			man, err := s.readManifest(sc.ID, v)
			if err != nil || !fileMetaEqual(currentMeta, man.Files) {
				continue
			}
			match := true
			for _, f := range man.Files {
				if f.Kind == FileKindRef {
					continue
				}
				snap, err := os.ReadFile(filepath.Join(s.versionPath(sc.ID, v.FileName), f.Name))
				if err != nil || !bytes.Equal(snap, locals[f.Name]) {
					match = false
					break
				}
			}
			if match {
				return v.ID
			}
			continue
		}
		if hasExtraLocal {
			continue
		}
		snap, err := os.ReadFile(s.versionPath(sc.ID, v.FileName))
		if err != nil {
			continue
		}
		if bytes.Equal(snap, locals[mainName]) {
			return v.ID
		}
	}
	return ""
}

// Get returns metadata and current workspace content.
func (s *Service) Get(id string) (*Detail, error) {
	all, err := s.loadAll()
	if err != nil {
		return nil, err
	}
	sc := findScriptIn(all, id)
	if sc == nil {
		return nil, ErrNotFound
	}
	return s.detailFor(*sc, all), nil
}

type CreateInput struct {
	Name          string
	Description   string
	Language      string
	Content       string
	InterpreterID string
	Files         []FileInput
	Variables     []Variable
}

// Create adds a new script with initial workspace files.
func (s *Service) Create(in CreateInput) (*Detail, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	prepared, err := PrepareFileInputs(in.Language, in.Content, in.Files)
	if err != nil {
		return nil, err
	}
	all, err := s.loadAll()
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	if err := validateRefsForLanguage(all, id, in.Language, prepared); err != nil {
		return nil, err
	}
	mainName, err := MainFileName(in.Language)
	if err != nil {
		return nil, err
	}
	if err := s.writeOwnedFiles(id, prepared); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	sc := Script{
		ID:            id,
		Name:          in.Name,
		Description:   in.Description,
		Language:      in.Language,
		InterpreterID: in.InterpreterID,
		FileName:      mainName,
		Files:         metaFromInputs(prepared),
		Variables:     NormalizeVariables(in.Variables),
		UpdatedAt:     now,
		Versions:      []Version{},
	}
	if err := s.store.Update(func(f *File) error {
		f.Scripts = append(f.Scripts, sc)
		return nil
	}); err != nil {
		return nil, err
	}
	all = append(all, sc)
	return s.detailFor(sc, all), nil
}

type UpdateInput struct {
	Name          string
	Description   string
	Language      string
	Content       string
	InterpreterID string
	Files         []FileInput
	Variables     []Variable
}

// Update saves metadata and workspace files.
func (s *Service) Update(id string, in UpdateInput) (*Detail, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	prepared, err := PrepareFileInputs(in.Language, in.Content, in.Files)
	if err != nil {
		return nil, err
	}
	mainName, err := MainFileName(in.Language)
	if err != nil {
		return nil, err
	}

	var updated Script
	var all []Script
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
		if err := validateRefsForLanguage(f.Scripts, id, in.Language, prepared); err != nil {
			return err
		}
		if err := s.writeOwnedFiles(id, prepared); err != nil {
			return err
		}
		sc := &f.Scripts[idx]
		sc.Name = in.Name
		sc.Description = in.Description
		sc.Language = in.Language
		sc.InterpreterID = in.InterpreterID
		sc.FileName = mainName
		sc.Files = metaFromInputs(prepared)
		sc.Variables = NormalizeVariables(in.Variables)
		sc.UpdatedAt = time.Now().UTC()
		updated = *sc
		all = append([]Script(nil), f.Scripts...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.detailFor(updated, all), nil
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

func (s *Service) findScript(id string) (*Script, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	sc := findScriptIn(f.Scripts, id)
	if sc == nil {
		return nil, ErrNotFound
	}
	cp := *sc
	return &cp, nil
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

// AddVersion copies the current workspace into versions/{name}/.
func (s *Service) AddVersion(id string, name, remark string) (*Version, error) {
	if name == "" {
		name = time.Now().UTC().Format("20060102150405")
	}
	if err := validateFileNameLoose(name); err != nil {
		return nil, err
	}
	sc, err := s.findScript(id)
	if err != nil {
		return nil, err
	}
	dstDir := s.versionPath(id, name)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return nil, err
	}
	meta := effectiveFiles(*sc)
	man := versionManifest{Files: meta}
	data, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dstDir, "manifest.json"), append(data, '\n'), 0o644); err != nil {
		return nil, err
	}
	for _, f := range meta {
		if f.Kind == FileKindRef {
			continue
		}
		text, err := s.readOwnedContent(*sc, f.Name)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dstDir, f.Name), []byte(text), 0o644); err != nil {
			return nil, err
		}
	}
	ver := Version{
		ID:        name,
		FileName:  name,
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
		_ = os.RemoveAll(dstDir)
		return nil, err
	}
	return &ver, nil
}

func validateFileNameLoose(name string) error {
	if name == "" || !fileNameRe.MatchString(name) || strings.Contains(name, "..") {
		return fmt.Errorf("invalid version id %q", name)
	}
	return nil
}

// UpdateVersionRemark updates the remark on an existing version snapshot.
func (s *Service) UpdateVersionRemark(id, versionID, remark string) (*Version, error) {
	var out Version
	err := s.store.Update(func(f *File) error {
		for i := range f.Scripts {
			if f.Scripts[i].ID != id {
				continue
			}
			for j := range f.Scripts[i].Versions {
				if f.Scripts[i].Versions[j].ID != versionID {
					continue
				}
				f.Scripts[i].Versions[j].Remark = remark
				out = f.Scripts[i].Versions[j]
				return nil
			}
			return ErrVersionNotFound
		}
		return ErrNotFound
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteVersion removes a version snapshot from metadata and deletes its files.
func (s *Service) DeleteVersion(id, versionID string) error {
	var fileName string
	err := s.store.Update(func(f *File) error {
		for i := range f.Scripts {
			if f.Scripts[i].ID != id {
				continue
			}
			sc := &f.Scripts[i]
			idx := -1
			for j := range sc.Versions {
				if sc.Versions[j].ID == versionID {
					idx = j
					fileName = sc.Versions[j].FileName
					break
				}
			}
			if idx < 0 {
				return ErrVersionNotFound
			}
			sc.Versions = append(sc.Versions[:idx], sc.Versions[idx+1:]...)
			return nil
		}
		return ErrNotFound
	})
	if err != nil {
		return err
	}
	if fileName == "" {
		return nil
	}
	p := s.versionPath(id, fileName)
	fi, statErr := os.Stat(p)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return nil
		}
		return statErr
	}
	if fi.IsDir() {
		if err := os.RemoveAll(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// Restore copies a version snapshot over the current workspace.
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
		fi, err := os.Stat(src)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			man, err := s.readManifest(id, *ver)
			if err != nil {
				return err
			}
			inputs := make([]FileInput, 0, len(man.Files))
			for _, mf := range man.Files {
				in := FileInput{
					Name:           mf.Name,
					Kind:           mf.Kind,
					SourceScriptID: mf.SourceScriptID,
					SourceFileName: mf.SourceFileName,
				}
				if mf.Kind != FileKindRef {
					b, err := os.ReadFile(filepath.Join(src, mf.Name))
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						return err
					}
					in.Content = string(b)
				}
				inputs = append(inputs, in)
			}
			if err := s.writeOwnedFiles(id, inputs); err != nil {
				return err
			}
			sc.Files = man.Files
			if mainName, err := MainFileName(sc.Language); err == nil {
				sc.FileName = mainName
			}
		} else {
			b, err := os.ReadFile(src)
			if err != nil {
				return err
			}
			mainName, err := MainFileName(sc.Language)
			if err != nil {
				return err
			}
			inputs := []FileInput{{Name: mainName, Kind: FileKindMain, Content: string(b)}}
			if err := s.writeOwnedFiles(id, inputs); err != nil {
				return err
			}
			sc.Files = metaFromInputs(inputs)
			sc.FileName = mainName
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
