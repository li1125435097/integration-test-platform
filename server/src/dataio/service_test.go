package dataio

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"integration-test-platform/server/src/executions"
	"integration-test-platform/server/src/kernelplans"
	"integration-test-platform/server/src/scripts"
	"integration-test-platform/server/src/store"
)

func TestImportMergeKeepsLocalAndReplacesSameID(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeScript(t, src, "overlap", "from-src", "src-body")
	writeScript(t, src, "fresh", "fresh-name", "fresh-body")
	writeVersion(t, src, "overlap", "v1", "snap")
	writeScript(t, dst, "overlap", "from-dst", "dst-body")
	writeScript(t, dst, "keep", "keep-name", "keep-body")
	writeVersion(t, dst, "overlap", "old", "gone")

	writeRecord(t, src, "rec-overlap", "overlap")
	writeRecord(t, src, "rec-new", "fresh")
	writeRecord(t, dst, "rec-overlap", "old-script")
	writeRecord(t, dst, "rec-keep", "keep")

	writePlan(t, src, "plan-overlap", "src-plan")
	writePlan(t, src, "plan-new", "new-plan")
	writePlan(t, dst, "plan-overlap", "dst-plan")
	writePlan(t, dst, "plan-keep", "keep-plan")

	writeFingerprint(t, src, "shared.txt", "new")
	writeFingerprint(t, src, "extra.txt", "extra")
	writeFingerprint(t, dst, "shared.txt", "old")
	writeFingerprint(t, dst, "only.txt", "stay")

	const localInterp = "{\"interpreters\":[{\"id\":\"local-box\"}]}\n"
	if err := os.WriteFile(filepath.Join(dst, "interpreters.json"), []byte(localInterp), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "interpreters.json"), []byte("{\"interpreters\":[{\"id\":\"other\"}]}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	payload := exportZip(t, src)
	names := zipNames(t, payload)
	for _, forbidden := range []string{
		"interpreters.json",
		"run-temp/secret.txt",
		"chrome_temp/secret.txt",
		"fingerprint-tmp/secret.txt",
	} {
		if containsName(names, forbidden) {
			t.Fatalf("zip contains %s", forbidden)
		}
	}
	for _, want := range []string{
		"scripts.json",
		"script-files/overlap/files/main.js",
		"script-files/overlap/versions/v1/manifest.json",
		"execution-records.json",
		"kernel-test-plans.json",
		"fingerprints/shared.txt",
	} {
		if !containsName(names, want) {
			t.Fatalf("zip missing %s in %v", want, names)
		}
	}

	result, err := NewService(dst).Import(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if result.Scripts.Added != 1 || result.Scripts.Updated != 1 {
		t.Fatalf("scripts counts = %+v", result.Scripts)
	}
	if result.Records.Added != 1 || result.Records.Updated != 1 {
		t.Fatalf("records counts = %+v", result.Records)
	}
	if result.Plans.Added != 1 || result.Plans.Updated != 1 {
		t.Fatalf("plans counts = %+v", result.Plans)
	}
	if result.Fingerprints.Added != 1 || result.Fingerprints.Updated != 1 {
		t.Fatalf("fingerprints counts = %+v", result.Fingerprints)
	}

	if got := readFile(t, filepath.Join(dst, "script-files", "overlap", "files", "main.js")); got != "src-body" {
		t.Fatalf("overlap body = %q", got)
	}
	if got := readFile(t, filepath.Join(dst, "script-files", "keep", "files", "main.js")); got != "keep-body" {
		t.Fatalf("keep body = %q", got)
	}
	if got := readFile(t, filepath.Join(dst, "script-files", "fresh", "files", "main.js")); got != "fresh-body" {
		t.Fatalf("fresh body = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dst, "script-files", "overlap", "versions", "old", "manifest.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old version still present: %v", err)
	}
	if got := readFile(t, filepath.Join(dst, "script-files", "overlap", "versions", "v1", "manifest.json")); got != "snap" {
		t.Fatalf("version = %q", got)
	}

	gotScripts, err := store.NewJSON[scripts.File](filepath.Join(dst, "scripts.json")).Load()
	if err != nil {
		t.Fatal(err)
	}
	if ids := scriptIDs(gotScripts.Scripts); stringsJoin(ids) != "overlap,keep,fresh" {
		t.Fatalf("script order = %v", ids)
	}

	gotRecords, err := store.NewJSON[executions.File](filepath.Join(dst, "execution-records.json")).Load()
	if err != nil {
		t.Fatal(err)
	}
	if ids := recordIDs(gotRecords.Records); stringsJoin(ids) != "rec-overlap,rec-keep,rec-new" {
		t.Fatalf("record order = %v", ids)
	}
	if gotRecords.Records[0].ScriptID != "overlap" {
		t.Fatalf("updated record = %+v", gotRecords.Records[0])
	}

	gotPlans, err := store.NewJSON[kernelplans.File](filepath.Join(dst, "kernel-test-plans.json")).Load()
	if err != nil {
		t.Fatal(err)
	}
	if ids := planIDs(gotPlans.Plans); stringsJoin(ids) != "plan-overlap,plan-keep,plan-new" {
		t.Fatalf("plan order = %v", ids)
	}
	if gotPlans.Plans[0].Name != "src-plan" {
		t.Fatalf("updated plan = %+v", gotPlans.Plans[0])
	}

	if got := readFile(t, filepath.Join(dst, "fingerprints", "shared.txt")); got != "new" {
		t.Fatalf("shared fingerprint = %q", got)
	}
	if got := readFile(t, filepath.Join(dst, "fingerprints", "only.txt")); got != "stay" {
		t.Fatalf("local fingerprint = %q", got)
	}
	if got := readFile(t, filepath.Join(dst, "fingerprints", "extra.txt")); got != "extra" {
		t.Fatalf("new fingerprint = %q", got)
	}
	if got := readFile(t, filepath.Join(dst, "interpreters.json")); got != localInterp {
		t.Fatalf("interpreters changed: %q", got)
	}
}

func TestExportSkipsTempFiles(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "s1", "one", "body")
	tmp := filepath.Join(dir, "script-files", "s1", "files", "main.js.tmp")
	if err := os.WriteFile(tmp, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{"run-temp", "chrome_temp", "fingerprint-tmp"} {
		p := filepath.Join(dir, sub, "secret.txt")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("secret"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "interpreters.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	names := zipNames(t, exportZip(t, dir))
	for _, name := range names {
		if name == "interpreters.json" || name == "script-files/s1/files/main.js.tmp" {
			t.Fatalf("exported %s", name)
		}
		for _, prefix := range []string{"run-temp/", "chrome_temp/", "fingerprint-tmp/"} {
			if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
				t.Fatalf("exported %s", name)
			}
		}
	}
	if !containsName(names, "script-files/s1/files/main.js") {
		t.Fatalf("missing script file: %v", names)
	}
}

func TestImportIgnoresInterpreterAndSkipsMissingFiles(t *testing.T) {
	parent := t.TempDir()
	dataDir := filepath.Join(parent, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const localInterp = "{\"interpreters\":[{\"id\":\"keep-me\"}]}\n"
	if err := os.WriteFile(filepath.Join(dataDir, "interpreters.json"), []byte(localInterp), 0o644); err != nil {
		t.Fatal(err)
	}
	writeRecord(t, dataDir, "rec-keep", "stay")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	addZip(t, zw, "interpreters.json", []byte("{\"interpreters\":[{\"id\":\"nope\"}]}"))
	addZip(t, zw, "scripts.json", []byte(`{"scripts":[{"id":"added","name":"n","language":"javascript","fileName":"main.js","updatedAt":"2020-01-01T00:00:00Z"}]}`))
	addZip(t, zw, "script-files/added/files/main.js", []byte("hello"))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	result, err := NewService(dataDir).Import(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if result.Scripts.Added != 1 || result.Records.Added != 0 || result.Records.Updated != 0 {
		t.Fatalf("result = %+v", result)
	}
	if got := readFile(t, filepath.Join(dataDir, "interpreters.json")); got != localInterp {
		t.Fatalf("interpreters = %q", got)
	}
	gotRecords, err := store.NewJSON[executions.File](filepath.Join(dataDir, "execution-records.json")).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(gotRecords.Records) != 1 || gotRecords.Records[0].ID != "rec-keep" {
		t.Fatalf("records = %+v", gotRecords.Records)
	}
	if got := readFile(t, filepath.Join(dataDir, "script-files", "added", "files", "main.js")); got != "hello" {
		t.Fatalf("imported script = %q", got)
	}
}

func TestImportRejectsPathTraversal(t *testing.T) {
	parent := t.TempDir()
	dataDir := filepath.Join(parent, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeScript(t, dataDir, "keep", "keep", "body")

	for _, name := range []string{"../evil.txt", "script-files/../../evil.txt", `script-files\..\evil.txt`} {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		addZip(t, zw, name, []byte("pwned"))
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		_, err := NewService(dataDir).Import(bytes.NewReader(buf.Bytes()))
		if !errors.Is(err, ErrInvalidArchive) {
			t.Fatalf("path %s: err = %v", name, err)
		}
		if _, statErr := os.Stat(filepath.Join(parent, "evil.txt")); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("path %s escaped: %v", name, statErr)
		}
	}
	if got := readFile(t, filepath.Join(dataDir, "script-files", "keep", "files", "main.js")); got != "body" {
		t.Fatalf("local script changed: %q", got)
	}
}

func exportZip(t *testing.T, dir string) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := NewService(dir).Export(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func zipNames(t *testing.T, payload []byte) []string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return names
}

func containsName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}

func addZip(t *testing.T, zw *zip.Writer, name string, body []byte) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(body); err != nil {
		t.Fatal(err)
	}
}

func writeScript(t *testing.T, dir, id, name, body string) {
	t.Helper()
	st := store.NewJSON[scripts.File](filepath.Join(dir, "scripts.json"))
	file, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	file.Scripts = append(file.Scripts, scripts.Script{
		ID:        id,
		Name:      name,
		Language:  "javascript",
		FileName:  "main.js",
		Files:     []scripts.ScriptFile{{Name: "main.js", Kind: scripts.FileKindMain}},
		UpdatedAt: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
	})
	if err := st.Save(file); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "script-files", id, "files", "main.js")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeVersion(t *testing.T, dir, id, version, body string) {
	t.Helper()
	path := filepath.Join(dir, "script-files", id, "versions", version, "manifest.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeRecord(t *testing.T, dir, id, scriptID string) {
	t.Helper()
	st := store.NewJSON[executions.File](filepath.Join(dir, "execution-records.json"))
	file, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	file.Records = append(file.Records, executions.Record{
		ID:        id,
		ScriptID:  scriptID,
		CreatedAt: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
	})
	if err := st.Save(file); err != nil {
		t.Fatal(err)
	}
}

func writePlan(t *testing.T, dir, id, name string) {
	t.Helper()
	st := store.NewJSON[kernelplans.File](filepath.Join(dir, "kernel-test-plans.json"))
	file, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	file.Plans = append(file.Plans, kernelplans.Plan{
		ID:        id,
		Name:      name,
		CreatedAt: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
	})
	if err := st.Save(file); err != nil {
		t.Fatal(err)
	}
}

func writeFingerprint(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, "fingerprints", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func scriptIDs(items []scripts.Script) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}

func recordIDs(items []executions.Record) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}

func planIDs(items []kernelplans.Plan) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}

func stringsJoin(parts []string) string {
	out := ""
	for i, part := range parts {
		if i > 0 {
			out += ","
		}
		out += part
	}
	return out
}
