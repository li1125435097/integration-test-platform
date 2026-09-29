package scripts

import "time"

// File kinds stored on a script workspace.
const (
	FileKindMain  = "main"
	FileKindLocal = "local"
	FileKindRef   = "ref"
)

// File is the on-disk scripts.json shape.
type File struct {
	Scripts []Script `json:"scripts"`
}

// ScriptFile is metadata for one tab in a script workspace.
type ScriptFile struct {
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	SourceScriptID string `json:"sourceScriptId,omitempty"`
	SourceFileName string `json:"sourceFileName,omitempty"`
}

// Variable is a {{name}} placeholder and its saved default value.
type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Script is metadata for one script; bodies live on disk under script-files/.
type Script struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Language      string       `json:"language"`
	InterpreterID string       `json:"interpreterId,omitempty"`
	FileName      string       `json:"fileName"`
	Files         []ScriptFile `json:"files,omitempty"`
	Variables     []Variable   `json:"variables,omitempty"`
	UpdatedAt     time.Time    `json:"updatedAt"`
	Versions      []Version    `json:"versions"`
}

// Version is a snapshot of the script workspace.
type Version struct {
	ID        string    `json:"id"`
	FileName  string    `json:"fileName"`
	Remark    string    `json:"remark,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// FileEntry is file metadata without content (list API / picker).
type FileEntry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// ListItem is returned by list API (no file content).
type ListItem struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Language       string      `json:"language"`
	InterpreterID  string      `json:"interpreterId,omitempty"`
	UpdatedAt      time.Time   `json:"updatedAt"`
	CurrentVersion string      `json:"currentVersion"`
	Files          []FileEntry `json:"files"`
	Variables      []Variable  `json:"variables,omitempty"`
}

// FileDetail is one editor tab including resolved content.
type FileDetail struct {
	Name             string `json:"name"`
	Kind             string `json:"kind"`
	Content          string `json:"content"`
	SourceScriptID   string `json:"sourceScriptId,omitempty"`
	SourceFileName   string `json:"sourceFileName,omitempty"`
	SourceScriptName string `json:"sourceScriptName,omitempty"`
	Missing          bool   `json:"missing,omitempty"`
}

// Detail includes script body for the editor.
type Detail struct {
	ListItem
	Content string       `json:"content"`
	Files   []FileDetail `json:"files"`
}

// FileInput is the create/update/preview payload for one workspace file.
type FileInput struct {
	Name           string
	Kind           string
	Content        string
	SourceScriptID string
	SourceFileName string
}

// WorkspaceFile is a materialized file ready to write into a run directory.
type WorkspaceFile struct {
	Name    string
	Content string
}
