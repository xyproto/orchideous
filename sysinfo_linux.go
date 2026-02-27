//go:build linux

package orchideous

func isLinux() bool  { return true }
func isDarwin() bool { return false }

const platformCDefine = "-D_GNU_SOURCE"

func extraLDLibPaths() []string { return nil }

func prependAsNeededFlag(ldflags []string) []string {
	return prependUnique(ldflags, "-Wl,--as-needed")
}
