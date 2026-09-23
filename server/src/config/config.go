package config

import (
	"flag"
	"os"
)

// App holds runtime settings from flags and environment.
type App struct {
	Addr      string
	ConfigDir string
	ShowVer   bool
}

// Parse reads CLI flags and applies environment overrides.
func Parse(version string) App {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	configDir := flag.String("config-dir", "", "Directory containing menu.json (default: ./config next to executable or project root)")
	showVer := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	cfg := App{
		Addr:      *addr,
		ConfigDir: *configDir,
		ShowVer:   *showVer,
	}
	if cfg.ShowVer {
		return cfg
	}
	if cfg.ConfigDir == "" {
		if v := os.Getenv("ITP_CONFIG_DIR"); v != "" {
			cfg.ConfigDir = v
		}
	}
	return cfg
}

func (a App) VersionString(version string) string {
	return version
}
