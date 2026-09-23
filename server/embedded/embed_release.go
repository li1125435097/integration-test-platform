//go:build release

package embedded

import "embed"

// Web contains synced frontend assets at build time (embedded/web).
//
//go:embed all:web
var Web embed.FS

// UseEmbed reports whether the binary was built with embedded web assets.
func UseEmbed() bool { return true }
