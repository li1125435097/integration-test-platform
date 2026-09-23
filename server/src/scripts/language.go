package scripts

import "fmt"

// ExtForLanguage maps script language to file extension (without dot).
func ExtForLanguage(lang string) (string, error) {
	switch lang {
	case "javascript":
		return "js", nil
	case "python":
		return "py", nil
	case "shell":
		return "sh", nil
	default:
		return "", fmt.Errorf("unsupported language %q", lang)
	}
}
