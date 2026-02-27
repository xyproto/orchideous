//go:build darwin

package orchideous

func isLinux() bool  { return false }
func isDarwin() bool { return true }

const platformCDefine = "-D_XOPEN_SOURCE=700"

func extraLDLibPaths() []string { return nil }

// macOS linker does not support --as-needed.
func prependAsNeededFlag(ldflags []string) []string { return ldflags }
