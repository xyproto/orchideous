//go:build windows

package orchideous

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// systemIncludeDirs returns the system include directories on Windows.
func systemIncludeDirs() []string {
	dirs := extraWindowsIncludeDirs()
	if fileExists("/usr/include") {
		dirs = append(dirs, "/usr/include")
	}
	cxx := findCompiler(false, false)
	if cxx != "" {
		out, err := exec.Command(cxx, "-dumpmachine").Output()
		if err == nil {
			machine := strings.TrimSpace(string(out))
			machineDir := "/usr/include/" + machine
			if fileExists(machineDir) {
				dirs = append(dirs, machineDir)
			}
		}
	}
	if fileExists("/usr/local/include") {
		dirs = append(dirs, "/usr/local/include")
	}
	if fileExists("/usr/pkg/include") {
		dirs = append(dirs, "/usr/pkg/include")
	}
	return dirs
}

// compilerSupportsStd checks if the compiler supports a given -std= flag.
// On Windows, uses a temp file instead of piping via sh -c.
func compilerSupportsStd(compiler, std string) bool {
	tmpFile := filepath.Join(os.TempDir(), "oh_stdcheck.cpp")
	os.WriteFile(tmpFile, []byte("int main(){}"), 0o644)
	defer os.Remove(tmpFile)
	cmd := exec.Command(compiler, "-std="+std, "-fsyntax-only", tmpFile)
	cmd.Stderr = nil
	cmd.Stdout = nil
	return cmd.Run() == nil
}
