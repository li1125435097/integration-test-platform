package interpreters

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type execMapping struct {
	language string
	names    []string
	version  func(path string) string
}

// Known interpreter executable basenames (checked under each PATH directory).
var execMappings = []execMapping{
	{language: "javascript", names: []string{"node", "nodejs"}, version: versionFlag("-v")},
	{language: "deno", names: []string{"deno"}, version: versionFlag("--version")},
	{language: "bun", names: []string{"bun"}, version: versionFlag("--version")},
	{language: "python", names: []string{"python", "python3", "py"}, version: versionCombined("--version")},
	{language: "shell", names: []string{"bash", "sh", "zsh", "fish", "pwsh"}, version: versionFlag("--version")},
	{language: "java", names: []string{"java"}, version: versionCombined("-version")},
	{language: "go", names: []string{"go"}, version: versionCombined("version")},
	{language: "rust", names: []string{"rustc", "cargo"}, version: versionCombined("--version")},
	{language: "ruby", names: []string{"ruby"}, version: versionCombined("--version")},
	{language: "php", names: []string{"php"}, version: versionCombined("-v")},
	{language: "perl", names: []string{"perl"}, version: versionCombined("-V")},
	{language: "lua", names: []string{"lua", "luajit"}, version: versionCombined("-v")},
	{language: "csharp", names: []string{"dotnet"}, version: versionCombined("--version")},
	{language: "kotlin", names: []string{"kotlin"}, version: versionCombined("-version")},
	{language: "scala", names: []string{"scala"}, version: versionCombined("-version")},
	{language: "dart", names: []string{"dart"}, version: versionCombined("--version")},
}

// envBinDirs adds <VAR>/bin (or subpath) when the env var is set.
var envBinDirs = []struct {
	env    string
	subdir string
}{
	{"JAVA_HOME", "bin"},
	{"GOROOT", "bin"},
	{"GRAALVM_HOME", "bin"},
	{"KOTLIN_HOME", "bin"},
	{"DART_SDK", "bin"},
	{"FLUTTER_ROOT", "bin/cache/dart-sdk/bin"},
}

// DetectVersion runs the version probe for language at path (first line only).
func DetectVersion(language, path string) string {
	for _, m := range execMappings {
		if m.language != language {
			continue
		}
		if m.version == nil {
			continue
		}
		if v := formatVersionLine(m.version(path)); v != "" {
			return v
		}
	}
	return ""
}

func formatVersionLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if idx := strings.IndexAny(s, "\r\n"); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}

// Discover scans PATH and common *_HOME bin directories for known interpreter executables.
func Discover(langFilter string) ([]DiscoverCandidate, error) {
	if langFilter != "" {
		if err := validateLanguage(langFilter); err != nil {
			return nil, err
		}
	}

	dirs := searchDirectories()
	seen := make(map[string]struct{})
	var out []DiscoverCandidate

	for _, dir := range dirs {
		for _, m := range execMappings {
			if langFilter != "" && m.language != langFilter {
				continue
			}
			for _, name := range m.names {
				for _, raw := range executablesInDir(dir, name) {
					abs, ok := absPath(raw)
					if !ok {
						continue
					}
					key := strings.ToLower(abs)
					if _, dup := seen[key]; dup {
						continue
					}
					seen[key] = struct{}{}
					out = append(out, DiscoverCandidate{
						Language:    m.language,
						Path:        abs,
						DefaultArgs: []string{},
						Version:     formatVersionLine(m.version(abs)),
					})
				}
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Language != out[j].Language {
			return out[i].Language < out[j].Language
		}
		return strings.ToLower(out[i].Path) < strings.ToLower(out[j].Path)
	})
	return out, nil
}

func searchDirectories() []string {
	var dirs []string
	if pathEnv := os.Getenv("PATH"); pathEnv != "" {
		dirs = append(dirs, filepath.SplitList(pathEnv)...)
	}
	for _, e := range envBinDirs {
		root := strings.TrimSpace(os.Getenv(e.env))
		if root == "" {
			continue
		}
		dirs = append(dirs, filepath.Join(root, filepath.FromSlash(e.subdir)))
	}
	return dedupeDirs(dirs)
}

func dedupeDirs(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, d := range in {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		abs, err := filepath.Abs(d)
		if err != nil {
			abs = d
		}
		key := strings.ToLower(abs)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, abs)
	}
	return out
}

func executablesInDir(dir, base string) []string {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil
	}
	if runtime.GOOS == "windows" {
		var paths []string
		for _, ext := range pathExts() {
			p := filepath.Join(dir, base+ext)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				paths = append(paths, p)
			}
		}
		return paths
	}
	p := filepath.Join(dir, base)
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return []string{p}
	}
	return nil
}

func pathExts() []string {
	raw := os.Getenv("PATHEXT")
	if raw == "" {
		return []string{".exe", ".com", ".bat", ".cmd"}
	}
	parts := strings.Split(strings.ToLower(raw), ";")
	var out []string
	seen := make(map[string]struct{})
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{".exe", ".com", ".bat", ".cmd"}
	}
	return out
}

func absPath(p string) (string, bool) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", false
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", false
	}
	if st, err := os.Stat(abs); err != nil || st.IsDir() {
		return "", false
	}
	return abs, true
}

func versionFlag(flag string) func(path string) string {
	return func(path string) string {
		out, err := exec.Command(path, flag).CombinedOutput()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
}

func versionCombined(flag string) func(path string) string {
	return func(path string) string {
		out, err := exec.Command(path, flag).CombinedOutput()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
}
