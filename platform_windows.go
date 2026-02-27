//go:build windows

package orchideous

func isLinux() bool  { return false }
func isDarwin() bool { return false }

const platformCDefine = "-D_XOPEN_SOURCE=700"

func extraLDLibPaths() []string { return nil }

// Windows does not use --as-needed.
func prependAsNeededFlag(ldflags []string) []string { return ldflags }

// detectPlatformType returns "generic" on Windows.
func detectPlatformType() string { return "generic" }
