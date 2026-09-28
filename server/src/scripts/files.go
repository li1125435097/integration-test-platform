package scripts

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	fileNameRe = regexp.MustCompile(`^[\w.-]+$`)

	// ErrRefMissing means a referenced file cannot be resolved.
	ErrRefMissing = errors.New("referenced file missing")
)

type versionManifest struct {
	Files []ScriptFile `json:"files"`
}

func (s *Service) filesDir(id string) string {
	return filepath.Join(s.scriptDir(id), "files")
}

func (s *Service) workspacePath(id, name string) string {
	return filepath.Join(s.filesDir(id), name)
}

func fileEntries(files []ScriptFile) []FileEntry {
	out := make([]FileEntry, 0, len(files))
	for _, f := range files {
		out = append(out, FileEntry{Name: f.Name, Kind: f.Kind})
	}
	return out
}

func effectiveFiles(sc Script) []ScriptFile {
	if len(sc.Files) > 0 {
		return sc.Files
	}
	mainName, err := MainFileName(sc.Language)
	if err != nil {
		mainName = "main.js"
	}
	return []ScriptFile{{Name: mainName, Kind: FileKindMain}}
}

func findScriptIn(all []Script, id string) *Script {
	for i := range all {
		if all[i].ID == id {
			return &all[i]
		}
	}
	return nil
}

func sourceHasFile(sc Script, name string) bool {
	for _, f := range effectiveFiles(sc) {
		if f.Name == name {
			return true
		}
	}
	return false
}

func validateFileName(name, lang string) error {
	if name == "" || !fileNameRe.MatchString(name) || strings.Contains(name, "..") {
		return fmt.Errorf("invalid file name %q", name)
	}
	ext, err := ExtForLanguage(lang)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(name, "."+ext) {
		return fmt.Errorf("file %q must end with .%s", name, ext)
	}
	return nil
}

func normalizeKind(in FileInput, mainName string) string {
	kind := strings.TrimSpace(in.Kind)
	if kind != "" {
		return kind
	}
	if in.SourceScriptID != "" {
		return FileKindRef
	}
	if in.Name == mainName {
		return FileKindMain
	}
	return FileKindLocal
}

// PrepareFileInputs validates workspace files. Empty files becomes a single main tab.
func PrepareFileInputs(language, content string, files []FileInput) ([]FileInput, error) {
	mainName, err := MainFileName(language)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		files = []FileInput{{Name: mainName, Kind: FileKindMain, Content: content}}
	}

	seen := make(map[string]struct{}, len(files))
	mainCount := 0
	out := make([]FileInput, 0, len(files))
	var mainFile FileInput
	var rest []FileInput

	for _, in := range files {
		in.Name = strings.TrimSpace(in.Name)
		in.Kind = normalizeKind(in, mainName)
		in.SourceScriptID = strings.TrimSpace(in.SourceScriptID)
		in.SourceFileName = strings.TrimSpace(in.SourceFileName)

		if err := validateFileName(in.Name, language); err != nil {
			return nil, err
		}
		if _, ok := seen[in.Name]; ok {
			return nil, fmt.Errorf("duplicate file name %q", in.Name)
		}
		seen[in.Name] = struct{}{}

		switch in.Kind {
		case FileKindMain:
			if in.Name != mainName {
				return nil, fmt.Errorf("main file must be named %s", mainName)
			}
			in.SourceScriptID = ""
			in.SourceFileName = ""
			mainCount++
			mainFile = in
		case FileKindLocal:
			if in.Name == mainName {
				return nil, fmt.Errorf("%s is reserved for the main file", mainName)
			}
			in.SourceScriptID = ""
			in.SourceFileName = ""
			rest = append(rest, in)
		case FileKindRef:
			if in.Name == mainName {
				return nil, fmt.Errorf("%s is reserved for the main file", mainName)
			}
			if in.SourceScriptID == "" || in.SourceFileName == "" {
				return nil, fmt.Errorf("file %q is a reference but source is missing", in.Name)
			}
			rest = append(rest, in)
		default:
			return nil, fmt.Errorf("unsupported file kind %q", in.Kind)
		}
	}
	if mainCount != 1 {
		return nil, fmt.Errorf("workspace must contain exactly one main file")
	}
	out = append(out, mainFile)
	out = append(out, rest...)
	return out, nil
}

func validateRefsForLanguage(all []Script, selfID, language string, files []FileInput) error {
	for _, in := range files {
		if in.Kind != FileKindRef {
			continue
		}
		if in.SourceScriptID == selfID {
			return fmt.Errorf("cannot reference a file from the same script")
		}
		src := findScriptIn(all, in.SourceScriptID)
		if src == nil {
			return fmt.Errorf("referenced script not found")
		}
		if src.Language != language {
			return fmt.Errorf("referenced script %q uses a different language", src.Name)
		}
		if err := validateFileName(in.SourceFileName, language); err != nil {
			return fmt.Errorf("referenced file %q: %w", in.SourceFileName, err)
		}
		if !sourceHasFile(*src, in.SourceFileName) {
			return fmt.Errorf("referenced file %q not found on script %q", in.SourceFileName, src.Name)
		}
	}
	return nil
}

func metaFromInputs(files []FileInput) []ScriptFile {
	out := make([]ScriptFile, 0, len(files))
	for _, in := range files {
		out = append(out, ScriptFile{
			Name:           in.Name,
			Kind:           in.Kind,
			SourceScriptID: in.SourceScriptID,
			SourceFileName: in.SourceFileName,
		})
	}
	return out
}

func (s *Service) loadAll() ([]Script, error) {
	f, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	return f.Scripts, nil
}

func (s *Service) readOwnedContent(sc Script, name string) (string, error) {
	b, err := os.ReadFile(s.workspacePath(sc.ID, name))
	if err == nil {
		return string(b), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	mainName, merr := MainFileName(sc.Language)
	if merr == nil && name == mainName && sc.FileName != "" {
		b, err = os.ReadFile(s.currentPath(sc.ID, sc.FileName))
		if err == nil {
			return string(b), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return "", nil
}

func (s *Service) resolveRefContent(all []Script, sourceID, sourceFileName string, visited map[string]bool) (string, error) {
	key := sourceID + "\x00" + sourceFileName
	if visited[key] {
		return "", fmt.Errorf("circular file reference")
	}
	visited[key] = true

	src := findScriptIn(all, sourceID)
	if src == nil {
		return "", fmt.Errorf("%w: script %s", ErrRefMissing, sourceID)
	}
	var meta *ScriptFile
	for i := range src.Files {
		if src.Files[i].Name == sourceFileName {
			meta = &src.Files[i]
			break
		}
	}
	if meta == nil {
		for _, f := range effectiveFiles(*src) {
			if f.Name == sourceFileName {
				cp := f
				meta = &cp
				break
			}
		}
	}
	if meta == nil {
		return "", fmt.Errorf("%w: %s", ErrRefMissing, sourceFileName)
	}
	if meta.Kind == FileKindRef {
		return s.resolveRefContent(all, meta.SourceScriptID, meta.SourceFileName, visited)
	}
	return s.readOwnedContent(*src, meta.Name)
}

func (s *Service) writeOwnedFiles(id string, files []FileInput) error {
	dir := s.filesDir(id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(s.scriptDir(id), "versions"), 0o755); err != nil {
		return err
	}
	want := make(map[string]struct{}, len(files))
	for _, in := range files {
		if in.Kind == FileKindRef {
			continue
		}
		if err := os.WriteFile(s.workspacePath(id, in.Name), []byte(in.Content), 0o644); err != nil {
			return err
		}
		want[in.Name] = struct{}{}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if _, ok := want[e.Name()]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	s.removeLegacyCurrent(id)
	return nil
}

func (s *Service) removeLegacyCurrent(id string) {
	for _, ext := range []string{"js", "py", "sh"} {
		_ = os.Remove(s.currentPath(id, "current."+ext))
	}
}

func (s *Service) inputsFromDisk(sc Script) ([]FileInput, error) {
	var inputs []FileInput
	for _, f := range effectiveFiles(sc) {
		in := FileInput{
			Name:           f.Name,
			Kind:           f.Kind,
			SourceScriptID: f.SourceScriptID,
			SourceFileName: f.SourceFileName,
		}
		if f.Kind != FileKindRef {
			content, err := s.readOwnedContent(sc, f.Name)
			if err != nil {
				return nil, err
			}
			in.Content = content
		}
		inputs = append(inputs, in)
	}
	return inputs, nil
}

// ResolveFiles materializes workspace files for execution. Refs are read from disk.
func (s *Service) ResolveFiles(selfID, language, content string, files []FileInput) (mainName string, out []WorkspaceFile, err error) {
	prepared, err := PrepareFileInputs(language, content, files)
	if err != nil {
		return "", nil, err
	}
	all, err := s.loadAll()
	if err != nil {
		return "", nil, err
	}
	if err := validateRefsForLanguage(all, selfID, language, prepared); err != nil {
		return "", nil, err
	}
	mainName, err = MainFileName(language)
	if err != nil {
		return "", nil, err
	}
	out = make([]WorkspaceFile, 0, len(prepared))
	for _, f := range prepared {
		wf := WorkspaceFile{Name: f.Name}
		if f.Kind == FileKindRef {
			text, rerr := s.resolveRefContent(all, f.SourceScriptID, f.SourceFileName, map[string]bool{})
			if rerr != nil {
				return "", nil, fmt.Errorf("引用文件 %s 不可用: %w", f.Name, rerr)
			}
			wf.Content = text
		} else {
			wf.Content = f.Content
		}
		out = append(out, wf)
	}
	return mainName, out, nil
}

// ResolveSaved materializes the on-disk workspace for a saved run.
func (s *Service) ResolveSaved(id string) (language, interpreterID, scriptName, mainName string, files []WorkspaceFile, err error) {
	sc, err := s.findScript(id)
	if err != nil {
		return "", "", "", "", nil, err
	}
	inputs, err := s.inputsFromDisk(*sc)
	if err != nil {
		return sc.Language, sc.InterpreterID, sc.Name, "", nil, err
	}
	mainName, files, err = s.ResolveFiles(id, sc.Language, "", inputs)
	if err != nil {
		return sc.Language, sc.InterpreterID, sc.Name, "", nil, err
	}
	return sc.Language, sc.InterpreterID, sc.Name, mainName, files, nil
}

func (s *Service) detailFor(sc Script, all []Script) *Detail {
	files := effectiveFiles(sc)
	details := make([]FileDetail, 0, len(files))
	var mainContent string
	mainName, _ := MainFileName(sc.Language)
	for _, f := range files {
		fd := FileDetail{
			Name:           f.Name,
			Kind:           f.Kind,
			SourceScriptID: f.SourceScriptID,
			SourceFileName: f.SourceFileName,
		}
		if f.Kind == FileKindRef {
			if src := findScriptIn(all, f.SourceScriptID); src != nil {
				fd.SourceScriptName = src.Name
			}
			text, err := s.resolveRefContent(all, f.SourceScriptID, f.SourceFileName, map[string]bool{})
			if err != nil {
				fd.Missing = true
			} else {
				fd.Content = text
			}
		} else {
			text, err := s.readOwnedContent(sc, f.Name)
			if err == nil {
				fd.Content = text
			}
		}
		if f.Kind == FileKindMain || f.Name == mainName {
			mainContent = fd.Content
		}
		details = append(details, fd)
	}
	item := s.listItemFor(sc)
	return &Detail{ListItem: item, Content: mainContent, Files: details}
}

func fileMetaEqual(a, b []ScriptFile) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Kind != b[i].Kind ||
			a[i].SourceScriptID != b[i].SourceScriptID || a[i].SourceFileName != b[i].SourceFileName {
			return false
		}
	}
	return true
}

func (s *Service) snapshotIsDir(id string, ver Version) bool {
	fi, err := os.Stat(s.versionPath(id, ver.FileName))
	return err == nil && fi.IsDir()
}

func (s *Service) readManifest(id string, ver Version) (versionManifest, error) {
	var man versionManifest
	data, err := os.ReadFile(filepath.Join(s.versionPath(id, ver.FileName), "manifest.json"))
	if err != nil {
		return man, err
	}
	err = json.Unmarshal(data, &man)
	return man, err
}

func (s *Service) localBytes(sc Script) (map[string][]byte, error) {
	out := make(map[string][]byte)
	for _, f := range effectiveFiles(sc) {
		if f.Kind == FileKindRef {
			continue
		}
		text, err := s.readOwnedContent(sc, f.Name)
		if err != nil {
			return nil, err
		}
		out[f.Name] = []byte(text)
	}
	return out, nil
}
