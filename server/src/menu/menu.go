package menu

import (
	"encoding/json"
	"fmt"
	"os"
)

// Item is one menu node (leaf or group). Leaf items use path (Vue Router hash path).
type Item struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Icon     string `json:"icon,omitempty"`
	Path     string `json:"path,omitempty"`
	Children []Item `json:"children,omitempty"`
}

// File is the on-disk menu.json shape.
type File struct {
	Items []Item `json:"items"`
}

// Load reads menu.json from path.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read menu config: %w", err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse menu config: %w", err)
	}
	return &f, nil
}

// Validate ensures menu shape is consistent (SPA routes, no HTML page files).
func (f *File) Validate() error {
	var walk func(items []Item, prefix string) error
	walk = func(items []Item, prefix string) error {
		for _, it := range items {
			label := prefix + it.ID
			if len(it.Children) > 0 {
				if it.Path != "" {
					return fmt.Errorf("menu item %q: group must not have path", label)
				}
				if err := walk(it.Children, label+"/"); err != nil {
					return err
				}
				continue
			}
			if it.Path == "" {
				return fmt.Errorf("menu item %q: leaf requires path", label)
			}
			if it.Path[0] != '/' {
				return fmt.Errorf("menu item %q: path must start with /", label)
			}
		}
		return nil
	}
	return walk(f.Items, "")
}
