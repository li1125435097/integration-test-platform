package dataio

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"integration-test-platform/server/src/executions"
	"integration-test-platform/server/src/kernelplans"
	"integration-test-platform/server/src/scripts"
	"integration-test-platform/server/src/store"
)

const (
	// MaxImportBytes is the maximum compressed upload and uncompressed total.
	MaxImportBytes int64 = 512 << 20
	maxZipEntries        = 100000
)

var (
	// ErrTooLarge means the upload or its uncompressed contents exceed MaxImportBytes.
	ErrTooLarge = errors.New("文件超过 512MB")
	// ErrInvalidArchive means the zip is not a usable data package.
	ErrInvalidArchive = errors.New("无效的数据包")
)

// Counts is how many records were inserted or replaced for one collection.
type Counts struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
}

// Result summarizes an import.
type Result struct {
	Scripts      Counts `json:"scripts"`
	Records      Counts `json:"records"`
	Plans        Counts `json:"plans"`
	Fingerprints Counts `json:"fingerprints"`
}

// Service reads and writes the application data directory.
type Service struct {
	dataDir string
}

// NewService exports and imports files under dataDir.
func NewService(dataDir string) *Service {
	return &Service{dataDir: dataDir}
}

// Export writes a zip of scripts, execution records, kernel plans, and fingerprints.
// Interpreter config and runtime temp directories are omitted.
func (s *Service) Export(w io.Writer) error {
	zw := zip.NewWriter(w)
	if err := addRegularFile(zw, s.dataDir, "scripts.json"); err != nil {
		_ = zw.Close()
		return err
	}
	if err := addTree(zw, s.dataDir, "script-files"); err != nil {
		_ = zw.Close()
		return err
	}
	if err := addRegularFile(zw, s.dataDir, "execution-records.json"); err != nil {
		_ = zw.Close()
		return err
	}
	if err := addRegularFile(zw, s.dataDir, "kernel-test-plans.json"); err != nil {
		_ = zw.Close()
		return err
	}
	if err := addTree(zw, s.dataDir, "fingerprints"); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

// Import merges a data zip into the local data directory.
// Same IDs are replaced, new IDs are appended, and local-only data is kept.
// interpreters.json in the archive is ignored.
func (s *Service) Import(r io.Reader) (Result, error) {
	var result Result
	payload, err := io.ReadAll(io.LimitReader(r, MaxImportBytes+1))
	if err != nil {
		return result, err
	}
	if int64(len(payload)) > MaxImportBytes {
		return result, ErrTooLarge
	}
	if !isZip(payload) {
		return result, fmt.Errorf("%w：不是 zip 文件", ErrInvalidArchive)
	}

	tmp, err := os.MkdirTemp("", "itp-import-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(tmp)

	if err := extractZip(bytes.NewReader(payload), int64(len(payload)), tmp); err != nil {
		return result, err
	}
	prepared, err := prepare(tmp)
	if err != nil {
		return result, err
	}
	if err := os.MkdirAll(s.dataDir, 0o755); err != nil {
		return result, err
	}
	result.Scripts, err = applyScripts(s.dataDir, tmp, prepared.scripts)
	if err != nil {
		return result, err
	}
	result.Records, err = applyRecords(s.dataDir, prepared.records)
	if err != nil {
		return result, err
	}
	result.Plans, err = applyPlans(s.dataDir, prepared.plans)
	if err != nil {
		return result, err
	}
	result.Fingerprints, err = applyFingerprints(s.dataDir, tmp, prepared.fingerprintNames)
	return result, err
}

type archiveContents struct {
	scripts          *scripts.File
	records          *executions.File
	plans            *kernelplans.File
	fingerprintNames []string
}

func prepare(root string) (archiveContents, error) {
	var out archiveContents
	scriptsFile, ok, err := loadArchiveJSON[scripts.File](filepath.Join(root, "scripts.json"))
	if err != nil {
		return out, err
	}
	if ok {
		seen := make(map[string]struct{}, len(scriptsFile.Scripts))
		for _, sc := range scriptsFile.Scripts {
			if !validID(sc.ID) {
				return out, fmt.Errorf("%w：脚本 ID 无效", ErrInvalidArchive)
			}
			if _, dup := seen[sc.ID]; dup {
				return out, fmt.Errorf("%w：脚本 ID 重复", ErrInvalidArchive)
			}
			seen[sc.ID] = struct{}{}
		}
		out.scripts = &scriptsFile
	}
	recordsFile, ok, err := loadArchiveJSON[executions.File](filepath.Join(root, "execution-records.json"))
	if err != nil {
		return out, err
	}
	if ok {
		if err := uniqueIDs(recordsFile.Records, func(r executions.Record) string { return r.ID }, "执行记录"); err != nil {
			return out, err
		}
		out.records = &recordsFile
	}
	plansFile, ok, err := loadArchiveJSON[kernelplans.File](filepath.Join(root, "kernel-test-plans.json"))
	if err != nil {
		return out, err
	}
	if ok {
		if err := uniqueIDs(plansFile.Plans, func(p kernelplans.Plan) string { return p.ID }, "内核方案"); err != nil {
			return out, err
		}
		out.plans = &plansFile
	}
	names, err := fingerprintNames(filepath.Join(root, "fingerprints"))
	if err != nil {
		return out, err
	}
	out.fingerprintNames = names
	return out, nil
}

func uniqueIDs[T any](items []T, idOf func(T) string, label string) error {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		id := idOf(item)
		if !validID(id) {
			return fmt.Errorf("%w：%s ID 无效", ErrInvalidArchive, label)
		}
		if _, dup := seen[id]; dup {
			return fmt.Errorf("%w：%s ID 重复", ErrInvalidArchive, label)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func fingerprintNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".tmp") {
			continue
		}
		if !validFingerprintName(name) {
			return nil, fmt.Errorf("%w：指纹文件名无效", ErrInvalidArchive)
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("%w：指纹文件无效", ErrInvalidArchive)
		}
		names = append(names, name)
	}
	return names, nil
}

func applyScripts(dataDir, extracted string, incoming *scripts.File) (Counts, error) {
	if incoming == nil {
		return Counts{}, nil
	}
	localPath := filepath.Join(dataDir, "scripts.json")
	local, err := store.NewJSON[scripts.File](localPath).Load()
	if err != nil {
		return Counts{}, err
	}
	merged, counts, err := mergeByID(local.Scripts, incoming.Scripts, func(sc scripts.Script) string { return sc.ID })
	if err != nil {
		return Counts{}, err
	}
	for _, sc := range incoming.Scripts {
		src := filepath.Join(extracted, "script-files", sc.ID)
		dst := filepath.Join(dataDir, "script-files", sc.ID)
		if err := replaceTree(dst, src); err != nil {
			return Counts{}, err
		}
	}
	if err := store.NewJSON[scripts.File](localPath).Save(scripts.File{Scripts: merged}); err != nil {
		return Counts{}, err
	}
	return counts, nil
}

func applyRecords(dataDir string, incoming *executions.File) (Counts, error) {
	if incoming == nil {
		return Counts{}, nil
	}
	return mergeJSON(
		filepath.Join(dataDir, "execution-records.json"),
		incoming.Records,
		func(f executions.File) []executions.Record { return f.Records },
		func(items []executions.Record) executions.File { return executions.File{Records: items} },
		func(r executions.Record) string { return r.ID },
	)
}

func applyPlans(dataDir string, incoming *kernelplans.File) (Counts, error) {
	if incoming == nil {
		return Counts{}, nil
	}
	return mergeJSON(
		filepath.Join(dataDir, "kernel-test-plans.json"),
		incoming.Plans,
		func(f kernelplans.File) []kernelplans.Plan { return f.Plans },
		func(items []kernelplans.Plan) kernelplans.File { return kernelplans.File{Plans: items} },
		func(p kernelplans.Plan) string { return p.ID },
	)
}

func applyFingerprints(dataDir, extracted string, names []string) (Counts, error) {
	if names == nil {
		return Counts{}, nil
	}
	dstDir := filepath.Join(dataDir, "fingerprints")
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return Counts{}, err
	}
	var counts Counts
	for _, name := range names {
		src := filepath.Join(extracted, "fingerprints", name)
		dst := filepath.Join(dstDir, name)
		_, err := os.Stat(dst)
		existed := err == nil
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return Counts{}, err
		}
		if err := copyFile(src, dst); err != nil {
			return Counts{}, err
		}
		if existed {
			counts.Updated++
		} else {
			counts.Added++
		}
	}
	return counts, nil
}

func mergeJSON[F any, T any](localPath string, incoming []T, items func(F) []T, wrap func([]T) F, idOf func(T) string) (Counts, error) {
	local, err := store.NewJSON[F](localPath).Load()
	if err != nil {
		return Counts{}, err
	}
	merged, counts, err := mergeByID(items(local), incoming, idOf)
	if err != nil {
		return Counts{}, err
	}
	if err := store.NewJSON[F](localPath).Save(wrap(merged)); err != nil {
		return Counts{}, err
	}
	return counts, nil
}

func mergeByID[T any](local, incoming []T, idOf func(T) string) ([]T, Counts, error) {
	incomingByID := make(map[string]T, len(incoming))
	order := make([]string, 0, len(incoming))
	for _, item := range incoming {
		id := idOf(item)
		if !validID(id) {
			return nil, Counts{}, fmt.Errorf("%w：ID 无效", ErrInvalidArchive)
		}
		if _, dup := incomingByID[id]; dup {
			return nil, Counts{}, fmt.Errorf("%w：ID 重复", ErrInvalidArchive)
		}
		incomingByID[id] = item
		order = append(order, id)
	}
	var counts Counts
	out := make([]T, 0, len(local)+len(incoming))
	for _, item := range local {
		id := idOf(item)
		if next, ok := incomingByID[id]; ok {
			out = append(out, next)
			delete(incomingByID, id)
			counts.Updated++
		} else {
			out = append(out, item)
		}
	}
	for _, id := range order {
		if item, ok := incomingByID[id]; ok {
			out = append(out, item)
			counts.Added++
		}
	}
	return out, counts, nil
}

func loadArchiveJSON[T any](path string) (T, bool, error) {
	var zero T
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return zero, false, nil
	}
	if err != nil {
		return zero, false, err
	}
	if err := json.Unmarshal(data, &zero); err != nil {
		return zero, false, fmt.Errorf("%w：%s", ErrInvalidArchive, filepath.Base(path))
	}
	return zero, true, nil
}

func addRegularFile(zw *zip.Writer, root, rel string) error {
	p := filepath.Join(root, filepath.FromSlash(rel))
	info, err := os.Lstat(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || strings.HasSuffix(info.Name(), ".tmp") {
		return nil
	}
	return addFile(zw, p, rel)
}

func addTree(zw *zip.Writer, root, rel string) error {
	dir := filepath.Join(root, filepath.FromSlash(rel))
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(d.Name(), ".tmp") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		relPath, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		return addFile(zw, p, filepath.ToSlash(relPath))
	})
}

func addFile(zw *zip.Writer, path, zipName string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = zipName
	hdr.Method = zip.Deflate
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func extractZip(r io.ReaderAt, size int64, dest string) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return fmt.Errorf("%w：不是 zip 文件", ErrInvalidArchive)
	}
	if len(zr.File) > maxZipEntries {
		return fmt.Errorf("%w：文件过多", ErrInvalidArchive)
	}
	var total uint64
	for _, f := range zr.File {
		rel, isDir, skip, err := classifyZipPath(f.Name)
		if err != nil {
			return err
		}
		if skip {
			continue
		}
		if f.FileInfo().Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w：不支持符号链接", ErrInvalidArchive)
		}
		if f.UncompressedSize64 > uint64(MaxImportBytes) {
			return ErrTooLarge
		}
		target, err := joinUnder(dest, rel)
		if err != nil {
			return err
		}
		if isDir || f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		n, err := extractFile(f, target)
		if err != nil {
			return err
		}
		if n > MaxImportBytes || total > uint64(MaxImportBytes)-uint64(n) {
			return ErrTooLarge
		}
		total += uint64(n)
	}
	return nil
}

func extractFile(f *zip.File, target string) (int64, error) {
	rc, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	n, err := io.Copy(out, io.LimitReader(rc, MaxImportBytes+1))
	if err != nil {
		return n, err
	}
	if n > MaxImportBytes {
		return n, ErrTooLarge
	}
	if err := out.Close(); err != nil {
		return n, err
	}
	return n, nil
}

func classifyZipPath(raw string) (rel string, isDir bool, skip bool, err error) {
	rel, isDir, err = safeZipPath(raw)
	if err != nil {
		return "", false, false, err
	}
	if strings.HasSuffix(path.Base(rel), ".tmp") {
		return rel, isDir, true, nil
	}
	switch {
	case rel == "scripts.json" || rel == "execution-records.json" || rel == "kernel-test-plans.json":
		return rel, isDir, false, nil
	case rel == "script-files" || strings.HasPrefix(rel, "script-files/"):
		return rel, isDir, false, nil
	case rel == "fingerprints" || strings.HasPrefix(rel, "fingerprints/"):
		return rel, isDir, false, nil
	default:
		return rel, isDir, true, nil
	}
}

func safeZipPath(name string) (string, bool, error) {
	name = strings.ReplaceAll(name, `\`, "/")
	if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "\x00") {
		return "", false, fmt.Errorf("%w：路径非法", ErrInvalidArchive)
	}
	if len(name) >= 2 && name[1] == ':' {
		return "", false, fmt.Errorf("%w：路径非法", ErrInvalidArchive)
	}
	isDir := strings.HasSuffix(name, "/")
	trimmed := strings.TrimSuffix(name, "/")
	if trimmed == "" || len(trimmed) > 512 {
		return "", false, fmt.Errorf("%w：路径非法", ErrInvalidArchive)
	}
	parts := strings.Split(trimmed, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false, fmt.Errorf("%w：路径非法", ErrInvalidArchive)
		}
	}
	return path.Join(parts...), isDir, nil
}

func joinUnder(root, rel string) (string, error) {
	root = filepath.Clean(root)
	dest := filepath.Join(root, filepath.FromSlash(rel))
	back, err := filepath.Rel(root, dest)
	if err != nil || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w：路径非法", ErrInvalidArchive)
	}
	return dest, nil
}

func replaceTree(dst, src string) error {
	info, err := os.Lstat(src)
	if errors.Is(err, os.ErrNotExist) {
		return os.RemoveAll(dst)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%w：脚本目录无效", ErrInvalidArchive)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	stage := dst + ".tmp"
	if err := os.RemoveAll(stage); err != nil {
		return err
	}
	if err := copyDir(src, stage); err != nil {
		_ = os.RemoveAll(stage)
		return err
	}
	if err := os.RemoveAll(dst); err != nil {
		_ = os.RemoveAll(stage)
		return err
	}
	if err := os.Rename(stage, dst); err != nil {
		return err
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(d.Name(), ".tmp") && p != src {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w：不支持符号链接", ErrInvalidArchive)
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%w：脚本文件无效", ErrInvalidArchive)
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func validID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}

func validFingerprintName(name string) bool {
	if name == "" || name != strings.TrimSpace(name) || name == "." || name == ".." {
		return false
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return false
	}
	return len(name) <= 255
}

func isZip(data []byte) bool {
	return len(data) >= 4 && data[0] == 'P' && data[1] == 'K'
}
