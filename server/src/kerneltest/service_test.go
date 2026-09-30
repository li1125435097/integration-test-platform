package kerneltest

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"integration-test-platform/server/src/fingerprints"
	"integration-test-platform/server/src/kernels"
	"integration-test-platform/server/src/scriptexec"
	"integration-test-platform/server/src/scripts"
)

type fakeRunner struct {
	mu       sync.Mutex
	inFlight int
	maxSeen  int
	calls    int
	delay    time.Duration
	lastArgs map[string]string
	argsByID []map[string]string
}

func (f *fakeRunner) RunSaved(id string, overrides map[string]string) *scriptexec.Output {
	f.mu.Lock()
	f.inFlight++
	f.calls++
	if f.inFlight > f.maxSeen {
		f.maxSeen = f.inFlight
	}
	copied := map[string]string{}
	for k, v := range overrides {
		copied[k] = v
	}
	f.lastArgs = copied
	f.argsByID = append(f.argsByID, copied)
	delay := f.delay
	f.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	f.mu.Lock()
	f.inFlight--
	f.mu.Unlock()
	return &scriptexec.Output{Success: true, ExitCode: 0, Stdout: overrides["KERNEL"]}
}

func (f *fakeRunner) snapshot() (calls, maxSeen int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls, f.maxSeen
}

func newTestService(t *testing.T, items []kernels.Kernel, runner scriptRunner) *Service {
	t.Helper()
	dir := t.TempDir()
	scriptSvc := scripts.NewService(dir)
	if err := scriptSvc.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}
	fp := fingerprints.NewService(dir, nil, nil)
	if err := fp.EnsureDir(); err != nil {
		t.Fatal(err)
	}
	return &Service{
		dataDir: dir,
		fp:      fp,
		scripts: scriptSvc,
		runner:  runner,
		scan: func() ([]kernels.Kernel, error) {
			return items, nil
		},
	}
}

func mustScript(t *testing.T, svc *Service, name string) string {
	t.Helper()
	detail, err := svc.scripts.Create(scripts.CreateInput{
		Name:     name,
		Language: "javascript",
		Content:  "console.log('{{KERNEL}}')",
	})
	if err != nil {
		t.Fatal(err)
	}
	return detail.ID
}

func mustKernel(t *testing.T, name string, withExe bool) kernels.Kernel {
	t.Helper()
	dir := t.TempDir()
	if withExe {
		if err := os.WriteFile(filepath.Join(dir, exeName), []byte("exe"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return kernels.Kernel{Name: name, Folder: name, Source: "dev", Path: dir}
}

func TestRunExePathArgsAndSharedFingerprint(t *testing.T) {
	first := mustKernel(t, "win152.26.14-dev", true)
	second := mustKernel(t, "win150.26.14-test", true)
	runner := &fakeRunner{}
	svc := newTestService(t, []kernels.Kernel{first, second}, runner)
	id := mustScript(t, svc, "内核测试-启动")
	cipher := []byte("cipher-bytes")

	out, err := svc.Run(Input{
		ScriptID:    id,
		Concurrency: 2,
		KernelPaths: []string{second.Path, first.Path},
		Args:        []string{" --disable-sync ", "", "--enable-automation"},
		Fingerprint: FingerprintInput{ContentBase64: base64.StdEncoding.EncodeToString(cipher)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Results) != 2 {
		t.Fatalf("results = %d", len(out.Results))
	}
	if out.Results[0].Kernel != second.Name || out.Results[1].Kernel != first.Name {
		t.Fatalf("order = %s, %s", out.Results[0].Kernel, out.Results[1].Kernel)
	}
	if out.Results[0].Variables.KERNEL != filepath.Join(second.Path, exeName) {
		t.Fatalf("KERNEL = %s", out.Results[0].Variables.KERNEL)
	}
	if out.Results[0].Variables.ARGS != "--disable-sync,--enable-automation" {
		t.Fatalf("ARGS = %q", out.Results[0].Variables.ARGS)
	}
	if out.Results[0].Variables.FINGERPRINT != out.Results[1].Variables.FINGERPRINT {
		t.Fatalf("fingerprint paths differ: %s vs %s", out.Results[0].Variables.FINGERPRINT, out.Results[1].Variables.FINGERPRINT)
	}
	got, err := os.ReadFile(out.Results[0].Variables.FINGERPRINT)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(cipher) {
		t.Fatalf("temp fingerprint = %q", got)
	}
	if filepath.Base(filepath.Dir(out.Results[0].Variables.FINGERPRINT)) != "fingerprint-tmp" {
		t.Fatalf("temp dir = %s", out.Results[0].Variables.FINGERPRINT)
	}
	saved, err := svc.fp.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 0 {
		t.Fatalf("saved fingerprints = %+v", saved)
	}
	if !out.Results[0].Success || out.Results[0].Stdout != jsString(filepath.Join(second.Path, exeName)) {
		t.Fatalf("first result = %+v", out.Results[0])
	}
	rawKernel := filepath.Join(second.Path, exeName)
	for _, overrides := range runner.argsByID {
		if overrides["KERNEL"] != jsString(rawKernel) && overrides["KERNEL"] != jsString(filepath.Join(first.Path, exeName)) {
			t.Fatalf("script KERNEL = %q", overrides["KERNEL"])
		}
		if !strings.Contains(overrides["FINGERPRINT"], `\\`) && strings.Contains(overrides["FINGERPRINT"], `\`) {
			t.Fatalf("fingerprint override was not escaped: %q", overrides["FINGERPRINT"])
		}
	}
}

func TestUserDirReused(t *testing.T) {
	item := mustKernel(t, "win152.26.14-dev", true)
	runner := &fakeRunner{}
	svc := newTestService(t, []kernels.Kernel{item}, runner)
	id := mustScript(t, svc, "内核测试-用户目录")
	in := Input{
		ScriptID:    id,
		Concurrency: 1,
		KernelPaths: []string{item.Path},
		Fingerprint: FingerprintInput{ContentBase64: base64.StdEncoding.EncodeToString([]byte("x"))},
	}
	first, err := svc.Run(in)
	if err != nil {
		t.Fatal(err)
	}
	dir := first.Results[0].Variables.USER_DIR
	want := filepath.Join(svc.dataDir, "chrome_temp", item.Name)
	abs, err := filepath.Abs(want)
	if err != nil {
		t.Fatal(err)
	}
	if dir != abs {
		t.Fatalf("USER_DIR = %s, want %s", dir, abs)
	}
	marker := filepath.Join(dir, "marker")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := svc.Run(in)
	if err != nil {
		t.Fatal(err)
	}
	if second.Results[0].Variables.USER_DIR != dir {
		t.Fatalf("reused USER_DIR = %s", second.Results[0].Variables.USER_DIR)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("marker: %v", err)
	}
}

func TestSavedFingerprintPath(t *testing.T) {
	item := mustKernel(t, "win104.11.6-dev", true)
	svc := newTestService(t, []kernels.Kernel{item}, &fakeRunner{})
	id := mustScript(t, svc, "内核测试-指纹")
	if err := svc.fp.Save("demo.bin", []byte("#BT@")); err != nil {
		t.Fatal(err)
	}
	out, err := svc.Run(Input{
		ScriptID:    id,
		KernelPaths: []string{item.Path},
		Fingerprint: FingerprintInput{SavedName: "demo.bin"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out.Results[0].Variables.FINGERPRINT)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "#BT@" {
		t.Fatalf("saved fingerprint = %q", got)
	}
	entries, err := os.ReadDir(filepath.Join(svc.dataDir, "fingerprint-tmp"))
	if err == nil && len(entries) != 0 {
		t.Fatalf("unexpected temp files: %d", len(entries))
	}
}

func TestMissingExeContinues(t *testing.T) {
	missing := mustKernel(t, "win1-dev", false)
	ok := mustKernel(t, "win2-dev", true)
	runner := &fakeRunner{}
	svc := newTestService(t, []kernels.Kernel{missing, ok}, runner)
	id := mustScript(t, svc, "内核测试-缺文件")
	out, err := svc.Run(Input{
		ScriptID:    id,
		Concurrency: 2,
		KernelPaths: []string{missing.Path, ok.Path},
		Fingerprint: FingerprintInput{ContentBase64: base64.StdEncoding.EncodeToString([]byte("x"))},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Results[0].Success || out.Results[0].Error != "未找到 BitBrowser.exe" {
		t.Fatalf("missing = %+v", out.Results[0])
	}
	if out.Results[0].Variables.KERNEL != filepath.Join(missing.Path, exeName) {
		t.Fatalf("KERNEL = %s", out.Results[0].Variables.KERNEL)
	}
	if !out.Results[1].Success {
		t.Fatalf("second = %+v", out.Results[1])
	}
	calls, _ := runner.snapshot()
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestScriptPrefixAndUnknownKernel(t *testing.T) {
	item := mustKernel(t, "win152.26.14-dev", true)
	svc := newTestService(t, []kernels.Kernel{item}, &fakeRunner{})
	plain := mustScript(t, svc, "普通脚本")
	_, err := svc.Run(Input{
		ScriptID:    plain,
		KernelPaths: []string{item.Path},
		Fingerprint: FingerprintInput{ContentBase64: base64.StdEncoding.EncodeToString([]byte("x"))},
	})
	if !errors.Is(err, ErrScriptPrefix) {
		t.Fatalf("prefix err = %v", err)
	}
	id := mustScript(t, svc, "内核测试-路径")
	_, err = svc.Run(Input{
		ScriptID:    id,
		KernelPaths: []string{filepath.Join(t.TempDir(), "nope")},
		Fingerprint: FingerprintInput{ContentBase64: base64.StdEncoding.EncodeToString([]byte("x"))},
	})
	if !errors.Is(err, ErrUnknownKernel) {
		t.Fatalf("unknown err = %v", err)
	}
}

func TestConcurrencyCap(t *testing.T) {
	items := []kernels.Kernel{
		mustKernel(t, "a-dev", true),
		mustKernel(t, "b-dev", true),
		mustKernel(t, "c-dev", true),
	}
	paths := []string{items[0].Path, items[1].Path, items[2].Path}
	runner := &fakeRunner{delay: 40 * time.Millisecond}
	svc := newTestService(t, items, runner)
	id := mustScript(t, svc, "内核测试-并发")
	in := Input{
		ScriptID:    id,
		Concurrency: 1,
		KernelPaths: paths,
		Fingerprint: FingerprintInput{ContentBase64: base64.StdEncoding.EncodeToString([]byte("x"))},
	}
	if _, err := svc.Run(in); err != nil {
		t.Fatal(err)
	}
	calls, maxSeen := runner.snapshot()
	if calls != 3 || maxSeen != 1 {
		t.Fatalf("serial calls=%d max=%d", calls, maxSeen)
	}

	runner = &fakeRunner{delay: 40 * time.Millisecond}
	svc.runner = runner
	in.Concurrency = 9
	if _, err := svc.Run(in); err != nil {
		t.Fatal(err)
	}
	calls, maxSeen = runner.snapshot()
	if calls != 3 || maxSeen != 3 {
		t.Fatalf("capped calls=%d max=%d", calls, maxSeen)
	}

	runner = &fakeRunner{delay: 40 * time.Millisecond}
	svc.runner = runner
	in.Concurrency = 2
	if _, err := svc.Run(in); err != nil {
		t.Fatal(err)
	}
	calls, maxSeen = runner.snapshot()
	if calls != 3 || maxSeen != 2 {
		t.Fatalf("parallel calls=%d max=%d", calls, maxSeen)
	}
}
