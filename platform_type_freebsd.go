//go:build freebsd

package orchideous

// detectPlatformType returns "freebsd" when pkg(8) is present.
func detectPlatformType() string {
	if fileExists("/usr/sbin/pkg") {
		return "freebsd"
	}
	return "generic"
}
