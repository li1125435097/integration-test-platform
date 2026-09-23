package assets

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"integration-test-platform/server/embedded"
)

// Web serves index, static, and pages from embed or disk.
type Web struct {
	root     http.FileSystem
	ioRoot   fs.FS
	pagesFS  fs.FS
	diskRoot string
}

// New builds a web asset provider for the current build mode.
func New() (*Web, error) {
	if embedded.UseEmbed() {
		sub, err := fs.Sub(embedded.Web, "web")
		if err != nil {
			return nil, err
		}
		pages, err := fs.Sub(sub, "pages")
		if err != nil {
			return nil, err
		}
		return &Web{
			root:    http.FS(sub),
			ioRoot:  sub,
			pagesFS: pages,
		}, nil
	}
	root, err := devWebDir()
	if err != nil {
		return nil, err
	}
	pagesDir := filepath.Join(root, "pages")
	return &Web{
		root:     http.Dir(root),
		pagesFS:  os.DirFS(pagesDir),
		diskRoot: root,
	}, nil
}

func devWebDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "web"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fs.ErrNotExist
		}
		dir = parent
	}
}

// RootFS is the web root filesystem (index.html, static/, pages/).
func (w *Web) RootFS() http.FileSystem {
	return w.root
}

// IORoot returns embed fs root for direct file reads (release builds).
func (w *Web) IORoot() fs.FS {
	return w.ioRoot
}

// SubFS returns an http.FileSystem for a subdirectory (static or pages).
func (w *Web) SubFS(name string) http.FileSystem {
	if w.diskRoot != "" {
		return http.Dir(filepath.Join(w.diskRoot, name))
	}
	sub, err := fs.Sub(w.ioRoot, name)
	if err != nil {
		panic("assets sub " + name + ": " + err.Error())
	}
	return http.FS(sub)
}

// DiskRoot returns the on-disk web directory in dev mode, or empty when embedded.
func (w *Web) DiskRoot() string {
	return w.diskRoot
}

// PageExists returns whether a page file exists under pages/.
func (w *Web) PageExists(name string) bool {
	if name == "" || filepath.Base(name) != name {
		return false
	}
	_, err := fs.Stat(w.pagesFS, name)
	return err == nil
}
