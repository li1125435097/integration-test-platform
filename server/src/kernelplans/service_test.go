package kernelplans

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestPlansCreateUpdateDelete(t *testing.T) {
	svc := NewService(t.TempDir())
	if err := svc.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Create("  ", Form{}); !errors.Is(err, ErrEmptyName) {
		t.Fatalf("empty name: %v", err)
	}

	first, err := svc.Create("  方案 A  ", Form{ScriptID: "s1", Kernels: []string{"k1"}})
	if err != nil {
		t.Fatal(err)
	}
	if first.Name != "方案 A" {
		t.Fatalf("name = %q", first.Name)
	}
	if first.Form.Sources == nil || first.Form.Flags == nil {
		t.Fatal("nil slices should be stored as empty arrays")
	}

	if _, err := svc.Create("方案 A", Form{}); !errors.Is(err, ErrNameExists) {
		t.Fatalf("duplicate: %v", err)
	}

	second, err := svc.Create("方案 B", Form{ScriptID: "s2"})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)
	updated, err := svc.Update(first.ID, "方案 A", Form{ScriptID: "s3", Kernels: []string{"k2"}})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Form.ScriptID != "s3" || !updated.CreatedAt.Equal(first.CreatedAt) {
		t.Fatalf("update = %+v", updated)
	}
	if !updated.UpdatedAt.After(first.UpdatedAt) {
		t.Fatal("updatedAt should move forward")
	}

	if _, err := svc.Update(second.ID, "方案 A", Form{}); !errors.Is(err, ErrNameExists) {
		t.Fatalf("rename conflict: %v", err)
	}
	if _, err := svc.Update("missing", "其它", Form{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing update: %v", err)
	}

	items, err := svc.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].ID != first.ID {
		t.Fatalf("list = %+v", items)
	}

	got, err := svc.Get(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Form.Kernels[0] != "k2" {
		t.Fatalf("get = %+v", got.Form)
	}

	if err := svc.Delete(first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(first.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted get: %v", err)
	}
	if err := svc.Delete(first.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete again: %v", err)
	}

	if _, err := os.Stat(svc.path); err != nil {
		t.Fatal(err)
	}
}
