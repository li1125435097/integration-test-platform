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

// MainFileName is the workspace entry file for a language (e.g. main.js).
func MainFileName(lang string) (string, error) {
	ext, err := ExtForLanguage(lang)
	if err != nil {
		return "", err
	}
	return "main." + ext, nil
}
