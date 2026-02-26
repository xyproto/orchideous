package main

import (
	"os"
	"strings"
	"testing"
)

func TestExtractLibs(t *testing.T) {
	ldflags := []string{"-lm", "-L/usr/lib", "-Wl,--as-needed", "-lpthread"}
	libs := extractLibs(ldflags)
	if len(libs) != 2 || libs[0] != "-lm" || libs[1] != "-lpthread" {
		t.Errorf("extractLibs = %v, want [-lm -lpthread]", libs)
	}
}

func TestExtractLinkFlags(t *testing.T) {
	ldflags := []string{"-lm", "-L/usr/lib", "-Wl,--as-needed"}
	flags := extractLinkFlags(ldflags)
	if len(flags) != 1 || flags[0] != "-Wl,--as-needed" {
		t.Errorf("extractLinkFlags = %v, want [-Wl,--as-needed]", flags)
	}
}

func TestFilterNonLinkFlags(t *testing.T) {
	flags := []string{"-O2", "-Wall", "-lm", "-L/usr/lib", "-Wl,--as-needed", "-I/usr/include"}
	result := filterNonLinkFlags(flags)
	for _, f := range result {
		if strings.HasPrefix(f, "-l") || strings.HasPrefix(f, "-L") || strings.HasPrefix(f, "-Wl,") {
			t.Errorf("filterNonLinkFlags should not contain %q", f)
		}
	}
	if len(result) != 3 {
		t.Errorf("expected 3 non-link flags, got %d: %v", len(result), result)
	}
}

func TestPkgNameFromInclude(t *testing.T) {
	cases := []struct {
		inc  string
		want string
	}{
		{"SDL2/SDL.h", "sdl2"},
		{"SDL2/SDL_image.h", "SDL2_image"},
		{"gtk/gtk.h", bestGtkPkg()},
		{"GL/glew.h", "glew"},
		{"GLFW/glfw3.h", "glfw3"},
		{"SFML/Graphics.hpp", "sfml-graphics"},
		{"boost/filesystem.hpp", ""},
		{"glm/vec3.hpp", "glm"},
		{"vulkan/vulkan.h", "vulkan"},
		{"raylib.h", "raylib"},
		{"QApplication", ""},
	}
	for _, tc := range cases {
		got := pkgNameFromInclude(tc.inc)
		if got != tc.want {
			t.Errorf("pkgNameFromInclude(%q) = %q, want %q", tc.inc, got, tc.want)
		}
	}
}

func TestDoCMake(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `#include <iostream>
int main() { std::cout << "hello"; return 0; }`)

	err := doCMake(BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("CMakeLists.txt")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "cmake_minimum_required") {
		t.Error("CMakeLists.txt missing cmake_minimum_required")
	}
	if !strings.Contains(content, "add_executable") {
		t.Error("CMakeLists.txt missing add_executable")
	}
	if !strings.Contains(content, "main.cpp") {
		t.Error("CMakeLists.txt missing main.cpp")
	}
}

func TestDoCMake_NoOverwrite(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)
	writeFile(t, "CMakeLists.txt", "existing")

	err := doCMake(BuildOptions{})
	if err == nil {
		t.Error("expected error when CMakeLists.txt already exists")
	}
}

func TestDoMakeFile(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `#include <iostream>
int main() { return 0; }`)

	err := doMakeFile()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, ".PHONY") {
		t.Error("Makefile missing .PHONY")
	}
	if !strings.Contains(content, "main.cpp") {
		t.Error("Makefile missing main.cpp reference")
	}
	if !strings.Contains(content, "clean:") {
		t.Error("Makefile missing clean target")
	}
}

func TestDoScript(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	err := doScript()
	if err != nil {
		t.Fatal(err)
	}

	buildSh, err := os.ReadFile("build.sh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buildSh), "#!/bin/sh") {
		t.Error("build.sh missing shebang")
	}
	if !strings.Contains(string(buildSh), "main.cpp") {
		t.Error("build.sh missing main.cpp")
	}

	cleanSh, err := os.ReadFile("clean.sh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cleanSh), "#!/bin/sh") {
		t.Error("clean.sh missing shebang")
	}
	if !strings.Contains(string(cleanSh), "rm -f") {
		t.Error("clean.sh missing rm command")
	}
}

func TestDoScript_NoOverwrite(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)
	writeFile(t, "build.sh", "existing")

	err := doScript()
	if err == nil {
		t.Error("expected error when build.sh already exists")
	}
}
