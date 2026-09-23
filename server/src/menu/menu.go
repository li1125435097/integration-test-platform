package menu

import (
	"encoding/json"
	"fmt"
	"os"
)

// Item is one menu node (leaf or group).
type Item struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Icon     string `json:"icon,omitempty"`
	Page     string `json:"page,omitempty"`
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

// Validate ensures every leaf page exists according to pageExists.
func (f *File) Validate(pageExists func(string) bool) error {
	var walk func(items []Item, path string) error
	walk = func(items []Item, prefix string) error {
		for _, it := range items {
			label := prefix + it.ID
			if len(it.Children) > 0 {
				if it.Page != "" {
					return fmt.Errorf("menu item %q: group must not have page", label)
				}
				if err := walk(it.Children, label+"/"); err != nil {
					return err
				}
				continue
			}
			if it.Page == "" {
				return fmt.Errorf("menu item %q: leaf requires page", label)
			}
			if !pageExists(it.Page) {
				return fmt.Errorf("menu item %q: page %q not found", label, it.Page)
			}
		}
		return nil
	}
	return walk(f.Items, "")
}
