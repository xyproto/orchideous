package orchideous

import (
	"os"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestAssembleFlags_DefaultBuild(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `#include <iostream>
int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{})

	if flags.Compiler == "" {
		t.Fatal("no compiler found")
	}
	if flags.Std == "" {
		t.Fatal("no std flag set")
	}
	assertFlagPresent(t, flags.CFlags, "-O2")
	assertFlagPresent(t, flags.CFlags, "-pipe")
	assertFlagPresent(t, flags.CFlags, "-fPIC")
	assertFlagPresent(t, flags.CFlags, "-Wall")
	assertFlagPresent(t, flags.CFlags, "-Wshadow")
	assertFlagPresent(t, flags.CFlags, "-Wpedantic")
}

func TestAssembleFlags_DebugBuild(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Debug: true})

	assertFlagPresent(t, flags.CFlags, "-O0")
	assertFlagPresent(t, flags.CFlags, "-g")
	assertFlagPresent(t, flags.CFlags, "-fno-omit-frame-pointer")
	assertFlagPresent(t, flags.LDFlags, "-fsanitize=address")
}

func TestAssembleFlags_DebugNoSan(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Debug: true, NoSanitizers: true})

	assertFlagPresent(t, flags.CFlags, "-O0")
	assertFlagAbsent(t, flags.LDFlags, "-fsanitize=address")
}

func TestAssembleFlags_OptBuild(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Opt: true})

	assertFlagPresent(t, flags.CFlags, "-Ofast")
	assertFlagPresent(t, flags.CFlags, "-flto")
}

func TestAssembleFlags_SmallBuild(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Small: true})

	assertFlagPresent(t, flags.CFlags, "-Os")
	assertFlagPresent(t, flags.CFlags, "-ffunction-sections")
	assertFlagPresent(t, flags.CFlags, "-fdata-sections")
	assertFlagAbsent(t, flags.CFlags, "-fPIC")
}

func TestAssembleFlags_TinyBuild(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Small: true, Tiny: true})

	assertFlagPresent(t, flags.CFlags, "-nostdlib")
	assertFlagPresent(t, flags.CFlags, "-fno-rtti")
}

func TestAssembleFlags_StrictBuild(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Strict: true})

	assertFlagPresent(t, flags.CFlags, "-Wextra")
	assertFlagPresent(t, flags.CFlags, "-Wconversion")
	assertFlagPresent(t, flags.CFlags, "-Weffc++")
}

func TestAssembleFlags_SloppyBuild(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Sloppy: true})

	assertFlagPresent(t, flags.CFlags, "-fpermissive")
	assertFlagPresent(t, flags.CFlags, "-w")
	assertFlagAbsent(t, flags.CFlags, "-Wall")
}

func TestAssembleFlags_CProject(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.c", `int main() { return 0; }`)

	proj := detectProject()
	if !proj.IsC {
		t.Fatal("expected IsC to be true")
	}
	flags := assembleFlags(proj, BuildOptions{})

	if runtime.GOOS == "linux" {
		if flags.Std != "c18" {
			t.Errorf("expected c18 for C on linux, got %q", flags.Std)
		}
	}
}

func TestAssembleFlags_OpenMP(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `#pragma omp parallel
int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{})

	assertFlagPresent(t, flags.CFlags, "-fopenmp")
	assertFlagPresent(t, flags.CFlags, "-O3")
	assertFlagPresent(t, flags.LDFlags, "-fopenmp")
	assertFlagPresent(t, flags.LDFlags, "-lpthread")
}

func TestAssembleFlags_LinuxHardening(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only test")
	}
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{})

	assertFlagPresent(t, flags.CFlags, "-fno-plt")
	assertFlagPresent(t, flags.CFlags, "-fstack-protector-strong")
}

func TestAssembleFlags_NoLinuxHardeningWhenSloppy(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only test")
	}
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Sloppy: true})

	assertFlagAbsent(t, flags.CFlags, "-fno-plt")
	assertFlagAbsent(t, flags.CFlags, "-fstack-protector-strong")
}

func TestAssembleFlags_ProfileGenerate(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Opt: true, ProfileGenerate: true})

	assertFlagPresent(t, flags.CFlags, "-fprofile-generate")
}

func TestAssembleFlags_ProfileUse(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{Opt: true, ProfileUse: true})

	assertFlagPresent(t, flags.CFlags, "-fprofile-use")
}

func TestAssembleFlags_CXXFLAGS(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	os.Setenv("CXXFLAGS", "-DTEST_FLAG -march=native")
	defer os.Unsetenv("CXXFLAGS")

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{})

	assertFlagPresent(t, flags.CFlags, "-DTEST_FLAG")
	assertFlagPresent(t, flags.CFlags, "-march=native")
}

func TestAssembleFlags_CFLAGS(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.c", `int main() { return 0; }`)

	os.Setenv("CFLAGS", "-DTEST_CFLAG -march=native")
	defer os.Unsetenv("CFLAGS")

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{})

	assertFlagPresent(t, flags.CFlags, "-DTEST_CFLAG")
	assertFlagPresent(t, flags.CFlags, "-march=native")
}

func TestAssembleFlags_LDFLAGS(t *testing.T) {
	withTempDir(t)
	writeFile(t, "main.cpp", `int main() { return 0; }`)

	os.Setenv("LDFLAGS", "-Wl,-z,relro,-z,now")
	defer os.Unsetenv("LDFLAGS")

	proj := detectProject()
	flags := assembleFlags(proj, BuildOptions{})

	assertFlagPresent(t, flags.LDFlags, "-Wl,-z,relro,-z,now")
}

func TestBuildCompileArgs(t *testing.T) {
	flags := BuildFlags{
		Compiler: "g++",
		Std:      "c++17",
		CFlags:   []string{"-O2", "-Wall"},
		LDFlags:  []string{"-lm"},
		Defines:  []string{"-DFOO"},
		IncPaths: []string{"include"},
	}
	args := buildCompileArgs(flags, []string{"main.cpp"}, "myapp")
	joined := strings.Join(args, " ")
	for _, want := range []string{"-std=c++17", "-O2", "-Wall", "-DFOO", "-Iinclude", "-o", "myapp", "main.cpp", "-lm"} {
		if !strings.Contains(joined, want) {
			t.Errorf("args missing %q: %s", want, joined)
		}
	}
}

func TestIsCompilerGCC(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/usr/bin/g++", true},
		{"/usr/bin/gcc", true},
		{"/usr/bin/x86_64-w64-mingw32-g++", true},
		{"/usr/bin/clang++", false},
		{"/usr/bin/cc", false},
	}
	for _, tc := range cases {
		if got := isCompilerGCC(tc.path); got != tc.want {
			t.Errorf("isCompilerGCC(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestIsCompilerClang(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/usr/bin/clang++", true},
		{"/usr/bin/clang", true},
		{"/usr/bin/g++", false},
		{"/usr/bin/cc", false},
	}
	for _, tc := range cases {
		if got := isCompilerClang(tc.path); got != tc.want {
			t.Errorf("isCompilerClang(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestDirDefines(t *testing.T) {
	withTempDir(t)
	os.MkdirAll("img", 0o755)
	os.MkdirAll("data", 0o755)

	defs := dirDefines()
	foundImg := false
	foundData := false
	for _, d := range defs {
		if strings.Contains(d, "IMGDIR") {
			foundImg = true
		}
		if strings.Contains(d, "DATADIR") {
			foundData = true
		}
	}
	if !foundImg {
		t.Error("expected IMGDIR define")
	}
	if !foundData {
		t.Error("expected DATADIR define")
	}
}

func TestMergeFlags(t *testing.T) {
	cflags := []string{"-O2"}
	ldflags := []string{}
	cflags, ldflags = mergeFlags(cflags, ldflags, "-I/usr/include -lm -L/usr/lib -Wl,--as-needed")
	assertFlagPresent(t, cflags, "-I/usr/include")
	assertFlagPresent(t, ldflags, "-lm")
	assertFlagPresent(t, ldflags, "-L/usr/lib")
	assertFlagPresent(t, ldflags, "-Wl,--as-needed")
}

func TestDotSlash(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"myapp", "./myapp"},
		{"./myapp", "./myapp"},
		{"/usr/bin/app", "/usr/bin/app"},
	}
	for _, tc := range cases {
		got := dotSlash(tc.input)
		if got != tc.want {
			t.Errorf("dotSlash(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// helpers
func assertFlagPresent(t *testing.T, flags []string, flag string) {
	t.Helper()
	if slices.Contains(flags, flag) {
		return
	}
	t.Errorf("expected flag %q in %v", flag, flags)
}

func assertFlagAbsent(t *testing.T, flags []string, flag string) {
	t.Helper()
	if slices.Contains(flags, flag) {
		t.Errorf("unexpected flag %q in %v", flag, flags)
		return
	}
}
