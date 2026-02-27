//go:build linux

package orchideous

// detectPlatformType returns the package-manager type for the current distro.
func detectPlatformType() string {
	// Arch Linux uses pacman; Debian/Ubuntu use dpkg.
	if fileExists("/usr/bin/pacman") {
		return "arch"
	}
	if fileExists("/usr/bin/dpkg-query") {
		return "deb"
	}
	return "generic"
}
