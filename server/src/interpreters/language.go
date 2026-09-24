package interpreters

import "fmt"

// SupportedLanguages are allowed language ids for stored interpreters.
var SupportedLanguages = map[string]struct{}{
	"javascript": {},
	"python":     {},
	"shell":      {},
	"java":       {},
	"go":         {},
	"rust":       {},
	"ruby":       {},
	"php":        {},
	"perl":       {},
	"lua":        {},
	"csharp":     {},
	"kotlin":     {},
	"scala":      {},
	"dart":       {},
	"deno":       {},
	"bun":        {},
}

func validateLanguage(lang string) error {
	if _, ok := SupportedLanguages[lang]; !ok {
		return fmt.Errorf("unsupported language %q", lang)
	}
	return nil
}
