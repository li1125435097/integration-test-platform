package scripts

import "time"

// File is the on-disk scripts.json shape.
type File struct {
	Scripts []Script `json:"scripts"`
}

// Script is metadata for one script; body lives on disk under script-files/.
type Script struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Language       string    `json:"language"`
	HasInterpreter bool      `json:"hasInterpreter"`
	FileName       string    `json:"fileName"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Versions       []Version `json:"versions"`
}

// Version is a snapshot of the script file.
type Version struct {
	ID        string    `json:"id"`
	FileName  string    `json:"fileName"`
	Remark    string    `json:"remark,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListItem is returned by list API (no file content).
type ListItem struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Language       string    `json:"language"`
	HasInterpreter bool      `json:"hasInterpreter"`
	UpdatedAt      time.Time `json:"updatedAt"`
	// CurrentVersion is the version id whose snapshot matches the current file.
	// Empty means the saved file differs from every snapshot (or there is none).
	CurrentVersion string `json:"currentVersion"`
}

// Detail includes script body for the editor.
type Detail struct {
	ListItem
	Content string `json:"content"`
}
