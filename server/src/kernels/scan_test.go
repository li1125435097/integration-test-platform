package kernels

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanLabelsAndOrder(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "BitBrowserDev", "Chrome-bin", "152", "win152.26.14", "Dictionaries"))
	mustMkdir(t, filepath.Join(root, "BitBrowserDev", "Chrome-bin", "104", "win104.11.6"))
	mustMkdir(t, filepath.Join(root, "BitBrowserTesting", "Chrome-bin", "152", "win152.26.14"))
	mustMkdir(t, filepath.Join(root, "BitBrowser", "Chrome-bin", "150", "win150.26.14"))
	if err := os.WriteFile(filepath.Join(root, "BitBrowserDev", "Chrome-bin", "readme.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"win152.26.14-dev",
		"win152.26.14-test",
		"win150.26.14-prod",
		"win104.11.6-dev",
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, kernels = %+v", len(got), names(got))
	}
	for i, name := range want {
		if got[i].Name != name {
			t.Fatalf("kernels[%d] = %s, want %s (%v)", i, got[i].Name, name, names(got))
		}
	}
	if got[0].Source != "dev" || got[0].Folder != "win152.26.14" {
		t.Fatalf("first kernel = %+v", got[0])
	}
	if got[2].Source != "prod" || got[2].Path == "" {
		t.Fatalf("prod kernel = %+v", got[2])
	}
}

func TestScanMissingRoots(t *testing.T) {
	root := t.TempDir()
	got, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d kernels", len(got))
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func names(items []Kernel) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.Name
	}
	return out
}
