//go:build freebsd

package orchideous

func isLinux() bool  { return false }
func isDarwin() bool { return false }

const platformCDefine = "-D_BSD_SOURCE"

func extraLDLibPaths() []string { return nil }

func prependAsNeededFlag(ldflags []string) []string {
	return prependUnique(ldflags, "-Wl,--as-needed")
}
