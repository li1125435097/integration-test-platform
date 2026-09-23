//go:build !release

package embedded

import "embed"

// Web is empty in development builds (disk serving from repo web/).
var Web embed.FS

// UseEmbed reports whether the binary was built with embedded web assets.
func UseEmbed() bool { return false }
