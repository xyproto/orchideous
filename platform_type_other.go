//go:build !linux && !darwin && !freebsd && !openbsd

package orchideous

// detectPlatformType returns "generic" on platforms without a known package manager.
func detectPlatformType() string {
	return "generic"
}
