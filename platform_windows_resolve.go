//go:build windows

package orchideous

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// msys2IncludePathToFlags resolves an include path to compiler/linker flags
// using MSYS2's pacman package manager.
func msys2IncludePathToFlags(includePath string) string {
	if includePath == "" || !fileExists(includePath) {
		return ""
	}
	pkg := lookupPackageOwnerMSYS2(includePath)
	if pkg == "" || skipPackages[pkg] {
		return ""
	}
	pcFiles := lookupPCFilesMSYS2(pkg)
	if len(pcFiles) == 0 {
		prefix := msys2Prefix()
		libDir := filepath.Join(prefix, "lib")
		result := tryLibFallbackWindows(includePath, pkg, []string{libDir})
		if result == "" && pkg != "boost" && pkg != "qt5-base" && pkg != "qt6-base" {
			fmt.Fprintf(os.Stderr, "WARNING: No pkg-config files for: %s\n", pkg)
		}
		return result
	}
	return pcFilesToFlags(pcFiles)
}

// lookupPackageOwnerMSYS2 queries pacman for the package owning a file.
func lookupPackageOwnerMSYS2(filePath string) string {
	// Convert Windows path to MSYS2-style path for pacman
	msysPath := windowsToMSYS2Path(filePath)
	out, err := exec.Command("pacman", "-Qo", "--quiet", msysPath).Output()
	if err != nil {
		// Try with the original Windows path
		out, err = exec.Command("pacman", "-Qo", "--quiet", filePath).Output()
		if err != nil {
			return ""
		}
	}
	return strings.TrimSpace(string(out))
}

// lookupPCFilesMSYS2 queries pacman for .pc files belonging to a package.
func lookupPCFilesMSYS2(pkg string) []string {
	if cached, ok := cachedPCFiles[pkg]; ok {
		return cached
	}
	out, err := exec.Command("pacman", "-Ql", pkg).Output()
	if err != nil {
		cachedPCFiles[pkg] = nil
		return nil
	}
	var pcFiles []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		// pacman -Ql output: "pkgname /path/to/file"
		parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(parts) == 2 && strings.HasSuffix(parts[1], ".pc") {
			pcFile := parts[1]
			// Convert MSYS2 path to Windows path
			if winPath := msys2ToWindowsPath(pcFile); winPath != "" {
				pcFiles = append(pcFiles, winPath)
			} else {
				pcFiles = append(pcFiles, pcFile)
			}
		}
	}
	cachedPCFiles[pkg] = pcFiles
	return pcFiles
}

// vcpkgIncludePathToFlags resolves an include path to flags using vcpkg's
// installed package tree.
func vcpkgIncludePathToFlags(includePath string) string {
	if includePath == "" || !fileExists(includePath) {
		return ""
	}
	installed := vcpkgInstalledDir()
	if installed == "" {
		return ""
	}
	// vcpkg provides pkg-config files in <installed>/lib/pkgconfig
	pkgconfigDir := filepath.Join(installed, "lib", "pkgconfig")
	if !fileExists(pkgconfigDir) {
		return ""
	}

	// Try to guess the package name from the include path
	pkgGuess := vcpkgGuessPackage(includePath, installed)
	if pkgGuess == "" {
		return ""
	}

	// Look for a .pc file matching the package guess
	pcFile := filepath.Join(pkgconfigDir, pkgGuess+".pc")
	if fileExists(pcFile) {
		return vcpkgPkgConfigFlags(pkgGuess, pkgconfigDir)
	}

	// Try to find any .pc file in the pkgconfig dir that matches
	matches, _ := filepath.Glob(filepath.Join(pkgconfigDir, "*.pc"))
	for _, m := range matches {
		pcName := strings.TrimSuffix(filepath.Base(m), ".pc")
		if strings.EqualFold(pcName, pkgGuess) {
			return vcpkgPkgConfigFlags(pcName, pkgconfigDir)
		}
	}

	// Fallback: try to link with -l<name> and -I/-L paths
	libDir := filepath.Join(installed, "lib")
	incDir := filepath.Join(installed, "include")
	return vcpkgLibFallback(pkgGuess, libDir, incDir)
}

// vcpkgGuessPackage guesses the vcpkg package name from an include path.
func vcpkgGuessPackage(includePath, installedDir string) string {
	incDir := filepath.Join(installedDir, "include")
	// Strip the include directory prefix to get the relative path
	rel, err := filepath.Rel(incDir, includePath)
	if err != nil {
		return ""
	}
	// The first path component is typically the package name
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) > 0 && parts[0] != "." && parts[0] != ".." {
		return strings.ToLower(parts[0])
	}
	return ""
}

// vcpkgPkgConfigFlags runs pkg-config with the vcpkg pkgconfig directory.
func vcpkgPkgConfigFlags(pkgName, pkgconfigDir string) string {
	cmd := exec.Command("pkg-config", "--cflags", "--libs", pkgName)
	cmd.Env = append(os.Environ(), "PKG_CONFIG_PATH="+pkgconfigDir)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// vcpkgLibFallback tries direct -l/-I/-L flags for a vcpkg package.
func vcpkgLibFallback(name, libDir, incDir string) string {
	// Check for .lib (MSVC) or .a (MinGW) files
	for _, ext := range []string{".lib", ".a"} {
		candidates := []string{
			filepath.Join(libDir, name+ext),
			filepath.Join(libDir, "lib"+name+ext),
		}
		for _, c := range candidates {
			if fileExists(c) {
				flags := "-l" + name
				if fileExists(incDir) {
					flags = "-I" + incDir + " " + flags
				}
				if fileExists(libDir) {
					flags += " -L" + libDir
				}
				return flags
			}
		}
	}
	return ""
}

// tryLibFallbackWindows is like tryLibFallback but checks for Windows library extensions.
func tryLibFallbackWindows(includePath, packageName string, libPaths []string) string {
	baseName := strings.TrimSuffix(filepath.Base(includePath), filepath.Ext(filepath.Base(includePath)))
	candidates := []string{packageName, baseName}

	for _, name := range candidates {
		if name == "" {
			continue
		}
		for _, libPath := range libPaths {
			for _, pattern := range []string{
				filepath.Join(libPath, "lib"+name+".a"),
				filepath.Join(libPath, "lib"+name+".dll.a"),
				filepath.Join(libPath, name+".lib"),
			} {
				if fileExists(pattern) {
					result := "-l" + name
					incDir := filepath.Dir(includePath)
					if fileExists(incDir) {
						result = "-I" + incDir + " " + result
					}
					result += " -L" + libPath
					return result
				}
			}
		}
	}
	return ""
}

// windowsToMSYS2Path converts a Windows path to an MSYS2 path.
func windowsToMSYS2Path(winPath string) string {
	// C:\msys64\mingw64\include\SDL2 -> /mingw64/include/SDL2
	winPath = filepath.ToSlash(winPath)
	// Strip common MSYS2 root prefixes
	for _, root := range []string{"C:/msys64", "C:/msys2", "D:/msys64"} {
		if strings.HasPrefix(strings.ToLower(winPath), strings.ToLower(root)) {
			return winPath[len(root):]
		}
	}
	return winPath
}

// msys2ToWindowsPath converts an MSYS2 path to a Windows path.
func msys2ToWindowsPath(msysPath string) string {
	if !strings.HasPrefix(msysPath, "/") {
		return msysPath
	}
	// Try to resolve via MSYS2 root
	for _, root := range []string{`C:\msys64`, `C:\msys2`, `D:\msys64`} {
		candidate := filepath.Join(root, filepath.FromSlash(msysPath))
		if fileExists(candidate) {
			return candidate
		}
	}
	// If MSYSTEM_PREFIX is set, use that as the base
	if prefix := os.Getenv("MSYSTEM_PREFIX"); prefix != "" {
		candidate := filepath.Join(filepath.Dir(filepath.Dir(prefix)), filepath.FromSlash(msysPath))
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}
