package fingerprints

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"integration-test-platform/server/src/scriptexec"
	"integration-test-platform/server/src/scripts"
)

const (
	encryptScriptName = "指纹加密"
	decryptScriptName = "指纹解密"
	encryptInputName  = "fingerprint-input.json"
	decryptInputName  = "fingerprint-input.bin"
	esmPackage        = "{\n  \"type\": \"module\"\n}\n"
)

var (
	// ErrScriptNotFound means the named encrypt or decrypt script is missing.
	ErrScriptNotFound = errors.New("未找到脚本")
	// ErrNotFound means a saved fingerprint file does not exist.
	ErrNotFound = errors.New("指纹文件不存在")
	// ErrInvalidName means the save name is empty or contains a path.
	ErrInvalidName = errors.New("文件名无效")
	// ErrInvalidInput means the plaintext or ciphertext payload cannot be used.
	ErrInvalidInput = errors.New("请求内容无效")
)

// FileInfo is one saved ciphertext file.
type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Updated time.Time `json:"updated"`
}

// Service encrypts and decrypts fingerprints through saved scripts, and stores ciphertext files.
type Service struct {
	dir     string
	scripts *scripts.Service
	runner  *scriptexec.Runner
}

// NewService stores files under dataDir/fingerprints.
func NewService(dataDir string, scriptSvc *scripts.Service, runner *scriptexec.Runner) *Service {
	return &Service{
		dir:     filepath.Join(dataDir, "fingerprints"),
		scripts: scriptSvc,
		runner:  runner,
	}
}

// EnsureDir creates the fingerprint directory.
func (s *Service) EnsureDir() error {
	return os.MkdirAll(s.dir, 0o755)
}

// Encrypt runs 指纹加密 and returns the ciphertext as standard Base64.
func (s *Service) Encrypt(plaintext, baseURL, bearer string) (string, error) {
	if strings.TrimSpace(plaintext) == "" {
		return "", fmt.Errorf("%w: 明文为空", ErrInvalidInput)
	}
	id, err := s.scriptID(encryptScriptName)
	if err != nil {
		return "", err
	}
	out := s.runner.RunSavedQuiet(id, map[string]string{
		"baseUrl": baseURL,
		"Bearer":  bearer,
	}, []scripts.WorkspaceFile{
		{Name: encryptInputName, Content: plaintext},
		{Name: "package.json", Content: esmPackage},
	})
	if out == nil || !out.Success {
		return "", runFailure(out)
	}
	encoded := strings.TrimSpace(out.Stdout)
	if _, err := base64.StdEncoding.DecodeString(encoded); err != nil || encoded == "" {
		return "", fmt.Errorf("%w: 加密结果不是合法 Base64", ErrInvalidInput)
	}
	return encoded, nil
}

// Decrypt runs 指纹解密 and returns pretty-printed JSON.
func (s *Service) Decrypt(contentBase64 string) (string, error) {
	raw, err := decodeCipher(contentBase64)
	if err != nil {
		return "", err
	}
	id, err := s.scriptID(decryptScriptName)
	if err != nil {
		return "", err
	}
	out := s.runner.RunSavedQuiet(id, nil, []scripts.WorkspaceFile{
		{Name: decryptInputName, Content: string(raw)},
		{Name: "package.json", Content: esmPackage},
	})
	if out == nil || !out.Success {
		return "", runFailure(out)
	}
	text := strings.TrimSpace(out.Stdout)
	if !json.Valid([]byte(text)) {
		return "", fmt.Errorf("%w: 解密结果不是合法 JSON", ErrInvalidInput)
	}
	return text, nil
}

// List returns saved ciphertext files, newest first.
func (s *Service) List() ([]FileInfo, error) {
	if err := s.EnsureDir(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	items := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		items = append(items, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			Updated: info.ModTime(),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].Updated.Equal(items[j].Updated) {
			return items[i].Updated.After(items[j].Updated)
		}
		return items[i].Name < items[j].Name
	})
	return items, nil
}

// Save writes ciphertext bytes. An existing name is overwritten.
func (s *Service) Save(name string, data []byte) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}
	p, err := s.filePath(name)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("%w: 密文为空", ErrInvalidInput)
	}
	return os.WriteFile(p, data, 0o644)
}

// Read returns the raw ciphertext bytes of a saved file.
func (s *Service) Read(name string) ([]byte, error) {
	p, err := s.filePath(name)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return data, err
}

func (s *Service) scriptID(name string) (string, error) {
	if s.scripts == nil {
		return "", fmt.Errorf("%w「%s」", ErrScriptNotFound, name)
	}
	items, err := s.scripts.List()
	if err != nil {
		return "", err
	}
	for _, item := range items {
		if item.Name == name {
			return item.ID, nil
		}
	}
	return "", fmt.Errorf("%w「%s」", ErrScriptNotFound, name)
}

func (s *Service) filePath(name string) (string, error) {
	if err := validateName(name); err != nil {
		return "", err
	}
	p := filepath.Join(s.dir, name)
	if filepath.Base(p) != name {
		return "", ErrInvalidName
	}
	return p, nil
}

func validateName(name string) error {
	if name == "" || name != strings.TrimSpace(name) || name == "." || name == ".." {
		return ErrInvalidName
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return ErrInvalidName
	}
	return nil
}

func decodeCipher(contentBase64 string) ([]byte, error) {
	encoded := strings.TrimSpace(contentBase64)
	if encoded == "" {
		return nil, fmt.Errorf("%w: 密文为空", ErrInvalidInput)
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: 密文不是合法 Base64", ErrInvalidInput)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: 密文为空", ErrInvalidInput)
	}
	return raw, nil
}

func runFailure(out *scriptexec.Output) error {
	if out == nil {
		return errors.New("脚本执行失败")
	}
	msg := strings.TrimSpace(out.Stderr)
	if msg == "" {
		msg = strings.TrimSpace(out.Error)
	}
	if msg == "" {
		msg = "脚本执行失败"
	}
	return errors.New(msg)
}
