package scripts

import (
	"regexp"
	"strings"
)

var (
	// {{name}} or {{name=default}}, with optional spaces around the name and equals sign.
	placeholderRe = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*)(?:\s*=\s*([^}]*?))?\s*\}\}`)
	varNameRe     = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// ExtractVariableNames returns placeholder names in first-seen order.
// {{name=value}} contributes the same name as {{name}}.
func ExtractVariableNames(contents ...string) []string {
	seen := map[string]bool{}
	var names []string
	for _, content := range contents {
		for _, sub := range placeholderRe.FindAllStringSubmatch(content, -1) {
			if len(sub) < 2 || seen[sub[1]] {
				continue
			}
			seen[sub[1]] = true
			names = append(names, sub[1])
		}
	}
	return names
}

// inlineDefault returns the trimmed {{name=value}} default and whether "=" was present.
func inlineDefault(match string, sub []string) (string, bool) {
	if !strings.Contains(match, "=") || len(sub) < 3 {
		return "", false
	}
	return strings.TrimSpace(sub[2]), true
}

// inlineDefaults collects the first {{name=value}} default for each name.
func inlineDefaults(contents ...string) map[string]string {
	seen := map[string]bool{}
	out := map[string]string{}
	for _, content := range contents {
		for _, sub := range placeholderRe.FindAllStringSubmatch(content, -1) {
			if len(sub) < 2 || seen[sub[1]] {
				continue
			}
			val, ok := inlineDefault(sub[0], sub)
			if !ok {
				continue
			}
			seen[sub[1]] = true
			out[sub[1]] = val
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// NormalizeVariables keeps the first value for each legal variable name.
func NormalizeVariables(in []Variable) []Variable {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]Variable, 0, len(in))
	for _, v := range in {
		name := v.Name
		if !varNameRe.MatchString(name) || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, Variable{Name: name, Value: v.Value})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// DefaultsMap indexes saved defaults by name.
func DefaultsMap(vars []Variable) map[string]string {
	if len(vars) == 0 {
		return nil
	}
	out := make(map[string]string, len(vars))
	for _, v := range vars {
		out[v.Name] = v.Value
	}
	return out
}

// ApplyVariables replaces {{name}} and {{name=value}} placeholders.
// A name present in overrides (including an empty string) wins over defaults.
// A name present in defaults wins over an inline {{name=value}} default.
// The first inline default for a name applies to every placeholder with that name.
// Otherwise the placeholder is replaced with an empty string.
func ApplyVariables(content string, defaults, overrides map[string]string) string {
	return applyPlaceholders(content, defaults, overrides, inlineDefaults(content))
}

func applyPlaceholders(content string, defaults, overrides, inline map[string]string) string {
	return placeholderRe.ReplaceAllStringFunc(content, func(match string) string {
		sub := placeholderRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		name := sub[1]
		if overrides != nil {
			if v, ok := overrides[name]; ok {
				return v
			}
		}
		if defaults != nil {
			if v, ok := defaults[name]; ok {
				return v
			}
		}
		if inline != nil {
			if v, ok := inline[name]; ok {
				return v
			}
		}
		return ""
	})
}

// ApplyFileVariables returns a copy of files with placeholders replaced.
// An inline default in any file applies to the same name in every file.
func ApplyFileVariables(files []WorkspaceFile, defaults, overrides map[string]string) []WorkspaceFile {
	if len(files) == 0 {
		return files
	}
	contents := make([]string, len(files))
	for i, f := range files {
		contents[i] = f.Content
	}
	inline := inlineDefaults(contents...)
	out := make([]WorkspaceFile, len(files))
	for i, f := range files {
		out[i] = WorkspaceFile{
			Name:    f.Name,
			Content: applyPlaceholders(f.Content, defaults, overrides, inline),
		}
	}
	return out
}
