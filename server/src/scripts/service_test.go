package scripts

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	dir := t.TempDir()
	s := NewService(dir)
	if err := s.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCreateGetMultiFile(t *testing.T) {
	s := newTestService(t)
	created, err := s.Create(CreateInput{
		Name:     "alpha",
		Language: "javascript",
		Files: []FileInput{
			{Name: "main.js", Kind: FileKindMain, Content: "require('./h')"},
			{Name: "h.js", Kind: FileKindLocal, Content: "module.exports = 1"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Content != "require('./h')" {
		t.Fatalf("content = %q", created.Content)
	}
	if len(created.Files) != 2 {
		t.Fatalf("files = %d", len(created.Files))
	}
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Files[1].Content != "module.exports = 1" {
		t.Fatalf("helper content = %q", got.Files[1].Content)
	}
	mainPath := filepath.Join(s.filesRoot, created.ID, "files", "main.js")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatal(err)
	}
}

func TestReferenceAndVersionRestore(t *testing.T) {
	s := newTestService(t)
	src, err := s.Create(CreateInput{
		Name:     "lib",
		Language: "javascript",
		Files: []FileInput{
			{Name: "main.js", Kind: FileKindMain, Content: "exports.x = 2"},
			{Name: "util.js", Kind: FileKindLocal, Content: "exports.y = 3"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	dst, err := s.Create(CreateInput{
		Name:     "app",
		Language: "javascript",
		Files: []FileInput{
			{Name: "main.js", Kind: FileKindMain, Content: "require('./u')"},
			{Name: "u.js", Kind: FileKindRef, SourceScriptID: src.ID, SourceFileName: "util.js"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if dst.Files[1].Content != "exports.y = 3" || dst.Files[1].Missing {
		t.Fatalf("ref not resolved: %+v", dst.Files[1])
	}

	ver, err := s.AddVersion(dst.ID, "v1", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(dst.ID, UpdateInput{
		Name:     "app",
		Language: "javascript",
		Files: []FileInput{
			{Name: "main.js", Kind: FileKindMain, Content: "changed"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	item, err := s.Restore(dst.ID, ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	if item.CurrentVersion != "v1" {
		t.Fatalf("current version = %q", item.CurrentVersion)
	}
	got, err := s.Get(dst.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "require('./u')" {
		t.Fatalf("restored main = %q", got.Content)
	}
	if len(got.Files) != 2 || got.Files[1].Kind != FileKindRef {
		t.Fatalf("restored files = %+v", got.Files)
	}
}

func TestLegacyCurrentFileRead(t *testing.T) {
	s := newTestService(t)
	id := "legacy-id"
	dir := filepath.Join(s.filesRoot, id)
	if err := os.MkdirAll(filepath.Join(dir, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "current.js"), []byte("legacy"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.store.Save(File{Scripts: []Script{{
		ID:       id,
		Name:     "legacy",
		Language: "javascript",
		FileName: "current.js",
	}}}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "legacy" {
		t.Fatalf("content = %q", got.Content)
	}
	if got.Files[0].Name != "main.js" {
		t.Fatalf("name = %q", got.Files[0].Name)
	}
}
