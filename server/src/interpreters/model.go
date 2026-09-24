package interpreters

import "time"

// File is the on-disk interpreters.json shape.
type File struct {
	Interpreters []Interpreter `json:"interpreters"`
}

// Interpreter is one configured language runtime.
type Interpreter struct {
	ID          string    `json:"id"`
	Language    string    `json:"language"`
	Path        string    `json:"path"`
	DefaultArgs []string  `json:"defaultArgs"`
	Version     string    `json:"version,omitempty"`
	IsDefault   bool      `json:"isDefault"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// DiscoverCandidate is a locally detected interpreter (not persisted).
type DiscoverCandidate struct {
	Language    string   `json:"language"`
	Path        string   `json:"path"`
	DefaultArgs []string `json:"defaultArgs"`
	Version     string   `json:"version,omitempty"`
}
