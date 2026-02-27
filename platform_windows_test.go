//go:build windows

package orchideous

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectPlatformType_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	// Without MSYSTEM or VCPKG_ROOT, should return "generic"
	origMSYSTEM := os.Getenv("MSYSTEM")
	origVCPKG := os.Getenv("VCPKG_ROOT")
	os.Unsetenv("MSYSTEM")
	os.Unsetenv("VCPKG_ROOT")
	defer func() {
		if origMSYSTEM != "" {
			os.Setenv("MSYSTEM", origMSYSTEM)
		}
		if origVCPKG != "" {
			os.Setenv("VCPKG_ROOT", origVCPKG)
		}
	}()

	pt := detectPlatformType()
	// Should be "generic" unless we're actually inside MSYS2 or have vcpkg
	if pt != "generic" && pt != "msys2" && pt != "vcpkg" {
		t.Errorf("unexpected platform type: %q", pt)
	}
}

func TestVcpkgRoot_Env(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	dir := t.TempDir()
	os.Setenv("VCPKG_ROOT", dir)
	defer os.Unsetenv("VCPKG_ROOT")

	got := vcpkgRoot()
	if got != dir {
		t.Errorf("vcpkgRoot() = %q, want %q", got, dir)
	}
}

func TestVcpkgTriplet_Default(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	os.Unsetenv("VCPKG_DEFAULT_TRIPLET")
	got := vcpkgTriplet()
	if got != "x64-windows" {
		t.Errorf("vcpkgTriplet() = %q, want %q", got, "x64-windows")
	}
}

func TestVcpkgTriplet_CustomEnv(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	os.Setenv("VCPKG_DEFAULT_TRIPLET", "x64-mingw-static")
	defer os.Unsetenv("VCPKG_DEFAULT_TRIPLET")

	got := vcpkgTriplet()
	if got != "x64-mingw-static" {
		t.Errorf("vcpkgTriplet() = %q, want %q", got, "x64-mingw-static")
	}
}

func TestVcpkgInstalledDir(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	root := t.TempDir()
	triplet := "x64-windows"
	installedDir := filepath.Join(root, "installed", triplet)
	os.MkdirAll(installedDir, 0o755)

	os.Setenv("VCPKG_ROOT", root)
	os.Unsetenv("VCPKG_DEFAULT_TRIPLET")
	defer os.Unsetenv("VCPKG_ROOT")

	got := vcpkgInstalledDir()
	if got != installedDir {
		t.Errorf("vcpkgInstalledDir() = %q, want %q", got, installedDir)
	}
}

func TestExtraWindowsIncludeDirs_Vcpkg(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	root := t.TempDir()
	triplet := "x64-windows"
	incDir := filepath.Join(root, "installed", triplet, "include")
	os.MkdirAll(incDir, 0o755)

	os.Setenv("VCPKG_ROOT", root)
	os.Unsetenv("VCPKG_DEFAULT_TRIPLET")
	os.Unsetenv("MSYSTEM")
	defer os.Unsetenv("VCPKG_ROOT")

	dirs := extraWindowsIncludeDirs()
	found := false
	for _, d := range dirs {
		if d == incDir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("extraWindowsIncludeDirs() = %v, expected to contain %q", dirs, incDir)
	}
}

func TestVcpkgGuessPackage(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	root := t.TempDir()
	incDir := filepath.Join(root, "include")
	os.MkdirAll(filepath.Join(incDir, "SDL2"), 0o755)
	writeFile(t, filepath.Join(incDir, "SDL2", "SDL.h"), "// SDL header")

	got := vcpkgGuessPackage(filepath.Join(incDir, "SDL2", "SDL.h"), root)
	if got != "sdl2" {
		t.Errorf("vcpkgGuessPackage() = %q, want %q", got, "sdl2")
	}
}

func TestVcpkgLibFallback(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	root := t.TempDir()
	libDir := filepath.Join(root, "lib")
	incDir := filepath.Join(root, "include")
	os.MkdirAll(libDir, 0o755)
	os.MkdirAll(incDir, 0o755)

	// Create a fake .lib file
	writeFile(t, filepath.Join(libDir, "SDL2.lib"), "fake lib")

	got := vcpkgLibFallback("SDL2", libDir, incDir)
	if got == "" {
		t.Error("vcpkgLibFallback() returned empty string, expected flags")
	}
}

func TestFindIncludeFile_CrossPlatform(t *testing.T) {
	dir := t.TempDir()
	// Create a nested include file
	subDir := filepath.Join(dir, "SDL2")
	os.MkdirAll(subDir, 0o755)
	writeFile(t, filepath.Join(subDir, "SDL.h"), "// SDL header")

	got := findIncludeFile(dir, "SDL2/SDL.h")
	if got == "" {
		t.Error("findIncludeFile() returned empty, expected to find SDL2/SDL.h")
	}
}

func TestCompilerSupportsStd_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	compiler := findCompiler(false, false)
	if compiler == "" {
		t.Skip("no C++ compiler found")
	}
	// c++17 should be widely supported
	if !compilerSupportsStd(compiler, "c++17") {
		t.Errorf("expected compiler %q to support c++17", compiler)
	}
}

func TestSystemIncludeDirs_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	// With a vcpkg root, systemIncludeDirs should include vcpkg paths
	root := t.TempDir()
	triplet := "x64-windows"
	incDir := filepath.Join(root, "installed", triplet, "include")
	os.MkdirAll(incDir, 0o755)

	os.Setenv("VCPKG_ROOT", root)
	os.Unsetenv("VCPKG_DEFAULT_TRIPLET")
	os.Unsetenv("MSYSTEM")
	defer os.Unsetenv("VCPKG_ROOT")

	dirs := systemIncludeDirs()
	found := false
	for _, d := range dirs {
		if d == incDir {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("systemIncludeDirs() = %v, expected to contain %q", dirs, incDir)
	}
}

func TestPlatformCDefine_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only test")
	}
	if platformCDefine != "-D_WIN32_WINNT=0x0601" {
		t.Errorf("platformCDefine = %q, want %q", platformCDefine, "-D_WIN32_WINNT=0x0601")
	}
}
