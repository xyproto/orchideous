//go:build openbsd

package orchideous

// detectPlatformType returns "openbsd" when pkg_info(1) is present.
func detectPlatformType() string {
	if fileExists("/usr/sbin/pkg_info") {
		return "openbsd"
	}
	return "generic"
}
