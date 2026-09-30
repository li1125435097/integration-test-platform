package interpreters

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBatchCreateSetsOneDefaultPerLanguage(t *testing.T) {
	dir := t.TempDir()
	py1 := writeFile(t, dir, "python.exe")
	py2 := writeFile(t, dir, "python3.exe")
	node := writeFile(t, dir, "node.exe")

	svc := NewService(dir)
	if err := svc.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}

	result, err := svc.BatchCreate([]BatchItem{
		{Language: "python", Path: py1, Version: "3.12"},
		{Language: "python", Path: py2, Version: "3.11"},
		{Language: "javascript", Path: node, Version: "v20"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Created) != 3 {
		t.Fatalf("created %d, want 3", len(result.Created))
	}
	if !result.Created[0].IsDefault || result.Created[0].Language != "python" {
		t.Fatalf("first python should be default: %+v", result.Created[0])
	}
	if result.Created[1].IsDefault {
		t.Fatalf("second python should not be default: %+v", result.Created[1])
	}
	if !result.Created[2].IsDefault || result.Created[2].Language != "javascript" {
		t.Fatalf("javascript should be default: %+v", result.Created[2])
	}

	goBin := writeFile(t, dir, "go.exe")
	again, err := svc.BatchCreate([]BatchItem{
		{Language: "python", Path: py2, Version: "3.11"},
		{Language: "go", Path: goBin, Version: "go1.22"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if again.Skipped != 1 {
		t.Fatalf("skipped %d, want 1", again.Skipped)
	}
	if len(again.Created) != 1 || !again.Created[0].IsDefault || again.Created[0].Language != "go" {
		t.Fatalf("new language should be default: %+v skipped=%d", again.Created, again.Skipped)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}
	defaults := map[string]int{}
	for _, it := range list {
		if it.IsDefault {
			defaults[it.Language]++
		}
	}
	if defaults["python"] != 1 || defaults["javascript"] != 1 || defaults["go"] != 1 {
		t.Fatalf("defaults = %+v", defaults)
	}
}

func TestBatchCreateKeepsExistingDefault(t *testing.T) {
	dir := t.TempDir()
	existing := writeFile(t, dir, "node.exe")
	extra := writeFile(t, dir, "nodejs.exe")

	svc := NewService(dir)
	now := time.Now().UTC()
	if err := svc.store.Save(File{Interpreters: []Interpreter{{
		ID:        "keep",
		Language:  "javascript",
		Path:      existing,
		IsDefault: true,
		UpdatedAt: now,
	}}}); err != nil {
		t.Fatal(err)
	}

	result, err := svc.BatchCreate([]BatchItem{
		{Language: "javascript", Path: extra, Version: "v18"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Created) != 1 || result.Created[0].IsDefault {
		t.Fatalf("existing default should stay: %+v", result.Created)
	}
}

func writeFile(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
