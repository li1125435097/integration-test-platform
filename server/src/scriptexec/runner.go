package scriptexec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"integration-test-platform/server/src/executions"
	"integration-test-platform/server/src/interpreters"
	"integration-test-platform/server/src/scripts"
)

const (
	maxCapture = 512 * 1024
	runTimeout = 2 * time.Minute
)

var ErrNoInterpreter = errors.New("no interpreter configured for this language")

// Runner executes scripts with configured interpreters.
type Runner struct {
	Scripts      *scripts.Service
	Interpreters *interpreters.Service
	Records      *executions.Service
	tempRoot     string
}

// NewRunner creates a runner; tempRoot holds short-lived script files.
func NewRunner(dataDir string, scriptsSvc *scripts.Service, interpSvc *interpreters.Service, recordsSvc *executions.Service) *Runner {
	return &Runner{
		Scripts:      scriptsSvc,
		Interpreters: interpSvc,
		Records:      recordsSvc,
		tempRoot:     filepath.Join(dataDir, "run-temp"),
	}
}

// PreviewInput runs editor content without persisting a record.
type PreviewInput struct {
	Language      string
	InterpreterID string
	Content       string
}

// Output is the result of a script run.
type Output struct {
	ExitCode   int    `json:"exitCode"`
	Success    bool   `json:"success"`
	DurationMs int64  `json:"durationMs"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	Error      string `json:"error,omitempty"`
	RecordID   string `json:"recordId,omitempty"`
}

// RunPreview executes content from the editor (no execution record).
func (r *Runner) RunPreview(in PreviewInput) *Output {
	if in.Language == "" {
		in.Language = "javascript"
	}
	return r.execute(in.Language, in.InterpreterID, in.Content, "", "", false)
}

// RunSaved executes the on-disk script and persists an execution record.
func (r *Runner) RunSaved(scriptID string) *Output {
	detail, err := r.Scripts.Get(scriptID)
	if err != nil {
		return &Output{Error: err.Error()}
	}
	return r.execute(detail.Language, detail.InterpreterID, detail.Content, scriptID, detail.Name, true)
}

func (r *Runner) execute(language, interpreterID, content, scriptID, scriptName string, persist bool) *Output {
	start := time.Now()
	out := &Output{}

	interp, err := r.resolveInterpreter(language, interpreterID)
	if err != nil {
		out.Error = err.Error()
		out.DurationMs = time.Since(start).Milliseconds()
		if persist {
			r.saveRecord(scriptID, scriptName, language, interpreterID, out, start)
		}
		return out
	}

	ext, err := scripts.ExtForLanguage(language)
	if err != nil {
		out.Error = err.Error()
		out.DurationMs = time.Since(start).Milliseconds()
		if persist {
			r.saveRecord(scriptID, scriptName, language, interpreterID, out, start)
		}
		return out
	}

	if err := os.MkdirAll(r.tempRoot, 0o755); err != nil {
		out.Error = err.Error()
		out.DurationMs = time.Since(start).Milliseconds()
		if persist {
			r.saveRecord(scriptID, scriptName, language, interpreterID, out, start)
		}
		return out
	}

	tmpPath := filepath.Join(r.tempRoot, uuid.NewString()+"."+ext)
	if err := os.WriteFile(tmpPath, []byte(content), 0o600); err != nil {
		out.Error = err.Error()
		out.DurationMs = time.Since(start).Milliseconds()
		if persist {
			r.saveRecord(scriptID, scriptName, language, interpreterID, out, start)
		}
		return out
	}
	defer func() { _ = os.Remove(tmpPath) }()

	args := append(append([]string{}, interp.DefaultArgs...), tmpPath)
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, interp.Path, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	out.DurationMs = time.Since(start).Milliseconds()
	out.Stdout = capOutput(stdout.String())
	out.Stderr = capOutput(stderr.String())

	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			out.ExitCode = exitErr.ExitCode()
		} else if errors.Is(runErr, context.DeadlineExceeded) {
			out.Error = fmt.Sprintf("执行超时（超过 %s）", runTimeout)
			out.ExitCode = -1
		} else {
			out.Error = runErr.Error()
			out.ExitCode = -1
		}
	} else {
		out.ExitCode = 0
	}
	out.Success = out.ExitCode == 0 && out.Error == ""

	if persist {
		rec := r.saveRecord(scriptID, scriptName, language, interpreterID, out, start)
		if rec != nil {
			out.RecordID = rec.ID
		}
	}
	return out
}

func (r *Runner) saveRecord(scriptID, scriptName, language, interpreterID string, out *Output, start time.Time) *executions.Record {
	if r.Records == nil {
		return nil
	}
	rec := executions.Record{
		ScriptID:      scriptID,
		ScriptName:    scriptName,
		Language:      language,
		InterpreterID: interpreterID,
		ExitCode:      out.ExitCode,
		Success:       out.Success,
		DurationMs:    out.DurationMs,
		Stdout:        out.Stdout,
		Stderr:        out.Stderr,
		Error:         out.Error,
		CreatedAt:     start.UTC(),
	}
	saved, err := r.Records.Add(rec)
	if err != nil {
		return nil
	}
	return saved
}

func (r *Runner) resolveInterpreter(language, id string) (*interpreters.Interpreter, error) {
	list, err := r.Interpreters.List()
	if err != nil {
		return nil, err
	}
	if id != "" {
		for i := range list {
			if list[i].ID != id {
				continue
			}
			if list[i].Language != language {
				return nil, fmt.Errorf("interpreter %q does not match language %q", id, language)
			}
			return &list[i], nil
		}
		return nil, fmt.Errorf("interpreter %q not found", id)
	}
	var fallback *interpreters.Interpreter
	for i := range list {
		if list[i].Language != language {
			continue
		}
		if list[i].IsDefault {
			return &list[i], nil
		}
		if fallback == nil {
			fallback = &list[i]
		}
	}
	if fallback != nil {
		return fallback, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrNoInterpreter, language)
}

func capOutput(s string) string {
	if len(s) <= maxCapture {
		return s
	}
	return s[:maxCapture] + "\n... (输出已截断)"
}

// FormatRunError returns a user-facing message when setup fails before run.
func FormatRunError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrNoInterpreter) {
		return "未配置该语言的解释器，请先在解释器管理中添加"
	}
	return strings.TrimSpace(err.Error())
}
