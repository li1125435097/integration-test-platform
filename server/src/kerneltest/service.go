package kerneltest

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"

	"integration-test-platform/server/src/fingerprints"
	"integration-test-platform/server/src/kernels"
	"integration-test-platform/server/src/scriptexec"
	"integration-test-platform/server/src/scripts"
)

const (
	scriptPrefix = "内核测试-"
	exeName      = "BitBrowser.exe"
)

var (
	// ErrNoScript means the request did not name a script.
	ErrNoScript = errors.New("请选择脚本")
	// ErrNoKernel means the request did not include a kernel.
	ErrNoKernel = errors.New("请选择内核")
	// ErrScriptPrefix means the script name is outside the kernel-test set.
	ErrScriptPrefix = errors.New("脚本名称必须以「内核测试-」开头")
	// ErrUnknownKernel means a path is not in the current kernel scan.
	ErrUnknownKernel = errors.New("内核不在扫描结果中")
	// ErrFingerprint means the fingerprint payload cannot be turned into a file.
	ErrFingerprint = errors.New("指纹内容无效")
)

// FingerprintInput is either a saved file or unsaved ciphertext/plaintext.
type FingerprintInput struct {
	SavedName     string
	ContentBase64 string
	Plaintext     string
	BaseURL       string
	Bearer        string
}

// Input is one kernel-test batch. Each kernel runs the script once.
type Input struct {
	ScriptID    string
	Concurrency int
	KernelPaths []string
	Args        []string
	Fingerprint FingerprintInput
}

// Variables are the placeholders substituted into the script.
type Variables struct {
	KERNEL      string `json:"KERNEL"`
	FINGERPRINT string `json:"FINGERPRINT"`
	ARGS        string `json:"ARGS"`
	USER_DIR    string `json:"USER_DIR"`
}

// Result is one kernel's script run.
type Result struct {
	Kernel     string    `json:"kernel"`
	Variables  Variables `json:"variables"`
	ExitCode   int       `json:"exitCode"`
	Success    bool      `json:"success"`
	DurationMs int64     `json:"durationMs"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	Error      string    `json:"error,omitempty"`
	RecordID   string    `json:"recordId,omitempty"`
}

// BatchOutput preserves the requested kernel order.
type BatchOutput struct {
	Results []Result `json:"results"`
}

type scriptRunner interface {
	RunSaved(scriptID string, overrides map[string]string) *scriptexec.Output
}

// Service prepares per-kernel variables and runs the selected script.
type Service struct {
	dataDir string
	fp      *fingerprints.Service
	scripts *scripts.Service
	runner  scriptRunner
	scan    func() ([]kernels.Kernel, error)
}

// NewService scans kernels with kernels.List and runs scripts through runner.
func NewService(dataDir string, fp *fingerprints.Service, scriptSvc *scripts.Service, runner *scriptexec.Runner) *Service {
	return &Service{
		dataDir: dataDir,
		fp:      fp,
		scripts: scriptSvc,
		runner:  runner,
		scan:    kernels.List,
	}
}

// Run resolves paths, then executes the script once per kernel with a concurrency limit.
func (s *Service) Run(in Input) (*BatchOutput, error) {
	if strings.TrimSpace(in.ScriptID) == "" {
		return nil, ErrNoScript
	}
	if len(in.KernelPaths) == 0 {
		return nil, ErrNoKernel
	}
	if s.scripts == nil {
		return nil, ErrNoScript
	}
	detail, err := s.scripts.Get(in.ScriptID)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(detail.Name, scriptPrefix) {
		return nil, ErrScriptPrefix
	}
	if s.scan == nil {
		return nil, errors.New("kernel scan is not configured")
	}
	scanned, err := s.scan()
	if err != nil {
		return nil, err
	}
	byPath := make(map[string]kernels.Kernel, len(scanned))
	for _, item := range scanned {
		byPath[filepath.Clean(item.Path)] = item
	}
	selected := make([]kernels.Kernel, 0, len(in.KernelPaths))
	for _, p := range in.KernelPaths {
		item, ok := byPath[filepath.Clean(p)]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownKernel, p)
		}
		selected = append(selected, item)
	}
	fpPath, err := s.fingerprintPath(in.Fingerprint)
	if err != nil {
		return nil, err
	}
	args := joinArgs(in.Args)
	concurrency := in.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(selected) {
		concurrency = len(selected)
	}

	results := make([]Result, len(selected))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, item := range selected {
		wg.Add(1)
		go func(i int, item kernels.Kernel) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = s.runOne(in.ScriptID, item, fpPath, args)
		}(i, item)
	}
	wg.Wait()
	return &BatchOutput{Results: results}, nil
}

func (s *Service) runOne(scriptID string, item kernels.Kernel, fpPath, args string) Result {
	exe := filepath.Join(item.Path, exeName)
	res := Result{
		Kernel: item.Name,
		Variables: Variables{
			KERNEL:      exe,
			FINGERPRINT: fpPath,
			ARGS:        args,
		},
		ExitCode: -1,
	}
	userDir, err := ensureUserDir(s.dataDir, item.Name)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Variables.USER_DIR = userDir
	info, err := os.Stat(exe)
	if err != nil || info.IsDir() {
		res.Error = "未找到 BitBrowser.exe"
		return res
	}
	if s.runner == nil {
		res.Error = "脚本执行器未配置"
		return res
	}
	out := s.runner.RunSaved(scriptID, map[string]string{
		"KERNEL":      jsString(exe),
		"FINGERPRINT": jsString(fpPath),
		"ARGS":        jsString(args),
		"USER_DIR":    jsString(userDir),
	})
	if out == nil {
		res.Error = "脚本执行失败"
		return res
	}
	res.ExitCode = out.ExitCode
	res.Success = out.Success
	res.DurationMs = out.DurationMs
	res.Stdout = out.Stdout
	res.Stderr = out.Stderr
	res.Error = out.Error
	res.RecordID = out.RecordID
	return res
}

func (s *Service) fingerprintPath(in FingerprintInput) (string, error) {
	if s.fp == nil {
		return "", fmt.Errorf("%w: 指纹服务未配置", ErrFingerprint)
	}
	if name := strings.TrimSpace(in.SavedName); name != "" {
		return s.fp.Path(name)
	}
	if encoded := strings.TrimSpace(in.ContentBase64); encoded != "" {
		raw, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil || len(raw) == 0 {
			return "", fmt.Errorf("%w: 密文不是合法 Base64", ErrFingerprint)
		}
		return s.writeTemp(raw)
	}
	if strings.TrimSpace(in.Plaintext) == "" {
		return "", fmt.Errorf("%w: 指纹内容为空", ErrFingerprint)
	}
	encoded, err := s.fp.Encrypt(in.Plaintext, in.BaseURL, in.Bearer)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(raw) == 0 {
		return "", fmt.Errorf("%w: 加密结果不是合法 Base64", ErrFingerprint)
	}
	return s.writeTemp(raw)
}

func (s *Service) writeTemp(data []byte) (string, error) {
	dir := filepath.Join(s.dataDir, "fingerprint-tmp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, uuid.NewString())
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func ensureUserDir(dataDir, name string) (string, error) {
	if name == "" || name != strings.TrimSpace(name) || name == "." || name == ".." || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", errors.New("内核名无效")
	}
	dir := filepath.Join(dataDir, "chrome_temp", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return abs, nil
}

// jsString keeps Windows paths intact inside '{{name}}' JavaScript string literals.
func jsString(s string) string {
	return strings.NewReplacer(`\`, `\\`, "\r", `\r`, "\n", `\n`).Replace(s)
}

func joinArgs(args []string) string {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		parts = append(parts, arg)
	}
	return strings.Join(parts, ",")
}
