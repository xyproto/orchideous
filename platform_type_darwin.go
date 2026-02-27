//go:build darwin

package orchideous

import "github.com/xyproto/files"

// detectPlatformType returns "brew" when Homebrew is available, else "generic".
func detectPlatformType() string {
	if files.WhichCached("brew") != "" {
		return "brew"
	}
	return "generic"
}
