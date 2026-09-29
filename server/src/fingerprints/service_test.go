package fingerprints

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"integration-test-platform/server/src/executions"
	"integration-test-platform/server/src/interpreters"
	"integration-test-platform/server/src/scriptexec"
	"integration-test-platform/server/src/scripts"
)

func TestSaveListRead(t *testing.T) {
	svc := NewService(t.TempDir(), nil, nil)
	if err := svc.Save("demo", []byte("#BT@")); err != nil {
		t.Fatal(err)
	}
	if err := svc.Save(`..\escape`, []byte("x")); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("path name err = %v", err)
	}
	if err := svc.Save("", []byte("x")); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("empty name err = %v", err)
	}

	items, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "demo" || items[0].Size != 4 {
		t.Fatalf("list = %+v", items)
	}
	got, err := svc.Read("demo")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("#BT@")) {
		t.Fatalf("read = %q", got)
	}
	if _, err := svc.Read("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing err = %v", err)
	}

	dir := filepath.Dir(svc.dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != "fingerprints" {
			t.Fatalf("unexpected data entry %s", entry.Name())
		}
	}
}

func TestEncryptMissingScript(t *testing.T) {
	dir := t.TempDir()
	scriptSvc := scripts.NewService(dir)
	if err := scriptSvc.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}
	interpSvc := interpreters.NewService(dir)
	if err := interpSvc.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}
	recordSvc := executions.NewService(dir)
	if err := recordSvc.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}
	runner := scriptexec.NewRunner(dir, scriptSvc, interpSvc, recordSvc)
	svc := NewService(dir, scriptSvc, runner)

	_, err := svc.Encrypt(`{"a":1}`, "https://example.test", "token")
	if !errors.Is(err, ErrScriptNotFound) {
		t.Fatalf("err = %v", err)
	}
	records, err := recordSvc.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("records = %+v", records)
	}
}
