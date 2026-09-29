package kernels

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Kernel is one Chrome-bin build folder, labeled with its install source.
type Kernel struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
	Source string `json:"source"`
	Path   string `json:"path"`
}

type sourceSpec struct {
	id     string
	dir    string
	suffix string
}

// Install roots under %AppData%\Roaming, each with Chrome-bin/<major>/<kernel>.
var sources = []sourceSpec{
	{id: "dev", dir: "BitBrowserDev", suffix: "dev"},
	{id: "test", dir: "BitBrowserTesting", suffix: "test"},
	{id: "prod", dir: "BitBrowser", suffix: "prod"},
}

var versionNumbers = regexp.MustCompile(`\d+`)

// AppDataDir is the roaming config directory (%AppData% on Windows).
func AppDataDir() (string, error) {
	if v := strings.TrimSpace(os.Getenv("APPDATA")); v != "" {
		return v, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate AppData: %w", err)
	}
	if strings.TrimSpace(dir) == "" {
		return "", fmt.Errorf("locate AppData: empty")
	}
	return dir, nil
}

// List scans the current user's BitBrowser kernel directories.
func List() ([]Kernel, error) {
	dir, err := AppDataDir()
	if err != nil {
		return nil, err
	}
	return Scan(dir)
}

// Scan reads Chrome-bin/<major>/<kernel> under each BitBrowser install root.
// Missing roots are skipped. Kernel names are "<folder>-dev|test|prod".
func Scan(appData string) ([]Kernel, error) {
	var out []Kernel
	for _, src := range sources {
		chromeBin := filepath.Join(appData, src.dir, "Chrome-bin")
		majors, err := os.ReadDir(chromeBin)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", chromeBin, err)
		}
		for _, major := range majors {
			if !major.IsDir() {
				continue
			}
			majorPath := filepath.Join(chromeBin, major.Name())
			entries, err := os.ReadDir(majorPath)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				folder := entry.Name()
				out = append(out, Kernel{
					Name:   folder + "-" + src.suffix,
					Folder: folder,
					Source: src.id,
					Path:   filepath.Join(majorPath, folder),
				})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		cmp := cmpVersion(out[i].Folder, out[j].Folder)
		if cmp != 0 {
			return cmp > 0
		}
		return sourceRank(out[i].Source) < sourceRank(out[j].Source)
	})
	return out, nil
}

func sourceRank(id string) int {
	switch id {
	case "dev":
		return 0
	case "test":
		return 1
	case "prod":
		return 2
	default:
		return 9
	}
}

func cmpVersion(a, b string) int {
	pa := versionParts(a)
	pb := versionParts(b)
	n := len(pa)
	if len(pb) < n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		if pa[i] != pb[i] {
			return pa[i] - pb[i]
		}
	}
	return len(pa) - len(pb)
}

func versionParts(name string) []int {
	raw := versionNumbers.FindAllString(name, -1)
	parts := make([]int, len(raw))
	for i, s := range raw {
		n, _ := strconv.Atoi(s)
		parts[i] = n
	}
	return parts
}
