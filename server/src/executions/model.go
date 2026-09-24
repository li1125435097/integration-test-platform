package executions

import "time"

// File is the on-disk execution-records.json shape.
type File struct {
	Records []Record `json:"records"`
}

// Record is one persisted script run from the script list.
type Record struct {
	ID            string    `json:"id"`
	ScriptID      string    `json:"scriptId"`
	ScriptName    string    `json:"scriptName"`
	Language      string    `json:"language"`
	InterpreterID string    `json:"interpreterId,omitempty"`
	ExitCode      int       `json:"exitCode"`
	Success       bool      `json:"success"`
	DurationMs    int64     `json:"durationMs"`
	Stdout        string    `json:"stdout"`
	Stderr        string    `json:"stderr"`
	Error         string    `json:"error,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}
