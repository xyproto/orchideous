//go:build windows

package orchideous

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func isLinux() bool  { return false }
func isDarwin() bool { return false }

// Windows doesn't need a POSIX source define; use an empty string so it's harmless if appended.
const platformCDefine = "-D_WIN32_WINNT=0x0601"

func extraLDLibPaths() []string { return nil }

// Windows does not use --as-needed.
func prependAsNeededFlag(ldflags []string) []string { return ldflags }

// detectPlatformType detects the Windows development environment.
// Returns "msys2" if running inside MSYS2/MinGW, "vcpkg" if vcpkg is
// available, or "generic" as a fallback.
func detectPlatformType() string {
	// MSYS2 sets MSYSTEM (MINGW64, UCRT64, CLANG64, etc.)
	if msystem := os.Getenv("MSYSTEM"); msystem != "" {
		if _, err := exec.LookPath("pacman"); err == nil {
			return "msys2"
		}
	}
	if vcpkgRoot() != "" {
		return "vcpkg"
	}
	return "generic"
}

// vcpkgRoot returns the vcpkg root directory, or "" if not found.
func vcpkgRoot() string {
	if root := os.Getenv("VCPKG_ROOT"); root != "" && fileExists(root) {
		return root
	}
	if p, err := exec.LookPath("vcpkg"); err == nil {
		return filepath.Dir(p)
	}
	return ""
}

// vcpkgTriplet returns the active vcpkg triplet for the current platform.
func vcpkgTriplet() string {
	if t := os.Getenv("VCPKG_DEFAULT_TRIPLET"); t != "" {
		return t
	}
	return "x64-windows"
}

// vcpkgInstalledDir returns the vcpkg installed directory for the active triplet.
func vcpkgInstalledDir() string {
	root := vcpkgRoot()
	if root == "" {
		return ""
	}
	triplet := vcpkgTriplet()
	dir := filepath.Join(root, "installed", triplet)
	if fileExists(dir) {
		return dir
	}
	return ""
}

// msys2Prefix returns the MSYS2 environment prefix (e.g. C:\msys64\mingw64).
func msys2Prefix() string {
	// MINGW_PREFIX is set by MSYS2 shells (e.g. /mingw64)
	if prefix := os.Getenv("MINGW_PREFIX"); prefix != "" {
		// Convert MSYS2 path to Windows path if needed
		if msysRoot := os.Getenv("MSYSTEM_PREFIX"); msysRoot != "" && fileExists(msysRoot) {
			return msysRoot
		}
		// Try common MSYS2 install locations
		for _, root := range []string{`C:\msys64`, `C:\msys2`, `D:\msys64`} {
			candidate := filepath.Join(root, filepath.FromSlash(prefix))
			if fileExists(candidate) {
				return candidate
			}
		}
	}
	// Fallback: infer from pacman location
	if p, err := exec.LookPath("pacman"); err == nil {
		// pacman is in <root>/usr/bin/pacman, and the mingw prefix is <root>/mingw64 etc.
		root := filepath.Dir(filepath.Dir(filepath.Dir(p)))
		msystem := strings.ToLower(os.Getenv("MSYSTEM"))
		switch {
		case strings.Contains(msystem, "clang64"):
			return filepath.Join(root, "clang64")
		case strings.Contains(msystem, "ucrt64"):
			return filepath.Join(root, "ucrt64")
		case strings.Contains(msystem, "clang32"):
			return filepath.Join(root, "clang32")
		case strings.Contains(msystem, "mingw32"):
			return filepath.Join(root, "mingw32")
		default: // MINGW64 or fallback
			return filepath.Join(root, "mingw64")
		}
	}
	return ""
}

// extraWindowsIncludeDirs returns additional include directories for the
// detected Windows development environment.
func extraWindowsIncludeDirs() []string {
	var dirs []string
	platform := detectPlatformType()
	switch platform {
	case "msys2":
		prefix := msys2Prefix()
		if prefix != "" {
			inc := filepath.Join(prefix, "include")
			if fileExists(inc) {
				dirs = append(dirs, inc)
			}
		}
	case "vcpkg":
		installed := vcpkgInstalledDir()
		if installed != "" {
			inc := filepath.Join(installed, "include")
			if fileExists(inc) {
				dirs = append(dirs, inc)
			}
		}
	}
	return dirs
}
