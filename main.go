package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const version = "3.3.3"

func main() {
	// Handle -C <dir> first, then re-dispatch
	args := os.Args[1:]
	if len(args) >= 2 && args[0] == "-C" {
		if err := os.Chdir(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		args = args[2:]
	}

	cmd := "build"
	if len(args) > 0 {
		cmd = args[0]
	}

	subArgs := args
	if len(subArgs) > 0 {
		subArgs = subArgs[1:]
	}

	switch cmd {
	case "-h", "--help", "help":
		printHelp()
	case "version", "--version":
		fmt.Printf("cbuild %s\n", version)
	case "build":
		exitOnErr(doBuild(BuildOptions{}))
	case "rebuild":
		doClean()
		exitOnErr(doBuild(BuildOptions{}))
	case "clean":
		doClean()
	case "fastclean":
		doFastClean()
	case "run":
		exitOnErr(doRun(BuildOptions{}, subArgs))
	case "debug":
		exitOnErr(doDebug(BuildOptions{Debug: true}, false))
	case "debugbuild":
		exitOnErr(doBuild(BuildOptions{Debug: true}))
	case "debugnosan":
		exitOnErr(doBuild(BuildOptions{Debug: true, NoSanitizers: true}))
	case "opt":
		exitOnErr(doBuild(BuildOptions{Opt: true}))
	case "strict":
		exitOnErr(doBuild(BuildOptions{Strict: true}))
	case "sloppy":
		exitOnErr(doBuild(BuildOptions{Sloppy: true}))
	case "small":
		exitOnErr(doBuild(BuildOptions{Small: true}))
	case "tiny":
		exitOnErr(doTiny(BuildOptions{Small: true, Tiny: true}))
	case "clang":
		exitOnErr(doBuild(BuildOptions{Clang: true}))
	case "clangdebug":
		exitOnErr(doDebug(BuildOptions{Clang: true, Debug: true}, true))
	case "clangstrict":
		exitOnErr(doBuild(BuildOptions{Clang: true, Strict: true}))
	case "clangsloppy":
		exitOnErr(doBuild(BuildOptions{Clang: true, Sloppy: true}))
	case "clangrebuild":
		doClean()
		exitOnErr(doBuild(BuildOptions{Clang: true}))
	case "clangtest":
		exitOnErr(doTest(BuildOptions{Clang: true}))
	case "test":
		exitOnErr(doTest(BuildOptions{}))
	case "testbuild":
		exitOnErr(doTestBuild(BuildOptions{}))
	case "rec":
		exitOnErr(doRec(subArgs))
	case "fmt":
		doFmt()
	case "cmake":
		if len(subArgs) > 0 && subArgs[0] == "ninja" {
			exitOnErr(doCMake(BuildOptions{}))
			exitOnErr(doNinja())
		} else {
			exitOnErr(doCMake(BuildOptions{}))
		}
	case "pro":
		exitOnErr(doPro(BuildOptions{}))
	case "ninja":
		exitOnErr(doNinja())
	case "ninja_install":
		exitOnErr(doNinjaInstall())
	case "ninja_clean":
		doNinjaClean()
	case "install":
		exitOnErr(doInstall())
	case "pkg":
		exitOnErr(doPkg())
	case "export":
		exitOnErr(doExport())
	case "make":
		exitOnErr(doMakeFile())
	case "script":
		exitOnErr(doScript())
	case "valgrind":
		exitOnErr(doValgrind(BuildOptions{}))
	case "win", "win64":
		exitOnErr(doBuild(BuildOptions{Win64: true}))
	case "smallwin", "smallwin64":
		exitOnErr(doBuild(BuildOptions{Win64: true, Small: true}))
	case "tinywin", "tinywin64":
		exitOnErr(doBuild(BuildOptions{Win64: true, Small: true, Tiny: true}))
	case "zap":
		exitOnErr(doBuild(BuildOptions{Zap: true}))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`cbuild %s

cbuild              - build the project
cbuild run          - build and run
cbuild debug        - debug build and launch debugger (gdb/cgdb)
cbuild debugbuild   - debug build (without launching debugger)
cbuild debugnosan   - debug build (without sanitizers)
cbuild opt          - optimized build
cbuild strict       - build with strict warning flags
cbuild sloppy       - build with sloppy flags
cbuild small        - build a smaller executable
cbuild tiny         - build a tiny executable (+ sstrip/upx)
cbuild clang        - build using clang++
cbuild clangdebug   - debug build using clang++ (launches lldb)
cbuild clangstrict  - use clang++ and strict flags
cbuild clangsloppy  - use clang++ and sloppy flags
cbuild clangrebuild - clean and build with clang++
cbuild clangtest    - build and run tests with clang++
cbuild clean        - remove built files
cbuild fastclean    - only remove executable and *.o
cbuild rebuild      - clean and build
cbuild test         - build and run tests
cbuild testbuild    - build tests (without running)
cbuild rec          - profile-guided optimization (build, run, rebuild)
cbuild fmt          - format source code with clang-format
cbuild cmake        - generate CMakeLists.txt
cbuild cmake ninja  - generate CMakeLists.txt and build with ninja
cbuild ninja        - build using existing CMakeLists.txt and ninja
cbuild ninja_install- install from ninja build
cbuild ninja_clean  - clean ninja build
cbuild pro          - generate QtCreator project file
cbuild install      - install the project (PREFIX, DESTDIR)
cbuild pkg          - package the project into pkg/
cbuild export       - export a standalone Makefile and build.sh
cbuild make         - generate a standalone Makefile
cbuild script       - generate build.sh and clean.sh
cbuild valgrind     - build and profile with valgrind
cbuild win64        - cross-compile for 64-bit Windows
cbuild smallwin64   - small win64 build
cbuild tinywin64    - tiny win64 build
cbuild zap          - build using zapcc++
cbuild version      - show version
cbuild -C <dir> ... - run in the given directory
`, version)
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func doRun(opts BuildOptions, runArgs []string) error {
	if err := doBuild(opts); err != nil {
		return err
	}
	exe := executableName()
	if exe == "" {
		return fmt.Errorf("no main source file found")
	}
	if opts.Win64 {
		exe += ".exe"
	}
	exePath := dotSlash(exe)
	c := exec.Command(exePath, runArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func doClean() {
	exe := executableName()
	patterns := []string{"*.o", "*.d", "common/*.o", "common/*.d", "include/*.o", "include/*.d", "*.profraw", "*.gcda", "*.gcno", ".sconsign.dblite", "callgrind.out.*"}
	for _, pat := range patterns {
		files, _ := filepath.Glob(pat)
		for _, f := range files {
			os.Remove(f)
			fmt.Println("Removed", f)
		}
	}
	if exe != "" {
		if err := os.Remove(exe); err == nil {
			fmt.Println("Removed", exe)
		}
		if err := os.Remove(exe + ".exe"); err == nil {
			fmt.Println("Removed", exe+".exe")
		}
	}
	// Clean test executables
	testSrcs := getTestSources()
	for _, ts := range testSrcs {
		testExe := strings.TrimSuffix(ts, filepath.Ext(ts))
		if err := os.Remove(testExe); err == nil {
			fmt.Println("Removed", testExe)
		}
	}
}

func doFastClean() {
	exe := executableName()
	files, _ := filepath.Glob("*.o")
	for _, f := range files {
		os.Remove(f)
		fmt.Println("Removed", f)
	}
	if exe != "" {
		if err := os.Remove(exe); err == nil {
			fmt.Println("Removed", exe)
		}
		if err := os.Remove(exe + ".exe"); err == nil {
			fmt.Println("Removed", exe+".exe")
		}
	}
}

func doTest(opts BuildOptions) error {
	testSrcs := getTestSources()
	if len(testSrcs) == 0 {
		fmt.Println("Nothing to test")
		return nil
	}

	proj := detectProject()
	flags := assembleFlags(proj, opts)
	for _, ts := range testSrcs {
		testExe := strings.TrimSuffix(ts, filepath.Ext(ts))
		srcs := append([]string{ts}, proj.DepSources...)
		if err := compileSources(srcs, testExe, flags); err != nil {
			return fmt.Errorf("building test %s: %w", testExe, err)
		}
		fmt.Printf("Running %s...\n", testExe)
		c := exec.Command(dotSlash(testExe))
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("test %s failed: %w", testExe, err)
		}
	}
	return nil
}

func doTestBuild(opts BuildOptions) error {
	proj := detectProject()
	flags := assembleFlags(proj, opts)

	// Build main if it exists
	if proj.MainSource != "" {
		exe := executableName()
		srcs := append([]string{proj.MainSource}, proj.DepSources...)
		if err := compileSources(srcs, exe, flags); err != nil {
			return err
		}
	}

	// Build all tests
	for _, ts := range proj.TestSources {
		testExe := strings.TrimSuffix(ts, filepath.Ext(ts))
		srcs := append([]string{ts}, proj.DepSources...)
		if err := compileSources(srcs, testExe, flags); err != nil {
			return fmt.Errorf("building test %s: %w", testExe, err)
		}
	}

	if proj.MainSource == "" && len(proj.TestSources) == 0 {
		fmt.Println("Nothing to build")
	}
	return nil
}

func doRec(runArgs []string) error {
	doClean()
	// Phase 1: Build with profile generation
	if err := doBuild(BuildOptions{Opt: true, ProfileGenerate: true}); err != nil {
		return fmt.Errorf("profile generation build: %w", err)
	}
	// Phase 2: Run to generate profile data
	exe := executableName()
	if exe == "" {
		return fmt.Errorf("no executable to run for profiling")
	}
	exePath := dotSlash(exe)
	c := exec.Command(exePath, runArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	_ = c.Run() // Don't fail if program exits with non-zero
	// Phase 3: Rebuild with profile use
	return doBuild(BuildOptions{Opt: true, ProfileUse: true})
}

func doFmt() {
	if _, err := exec.LookPath("clang-format"); err != nil {
		fmt.Fprintln(os.Stderr, "error: clang-format not found in PATH")
		os.Exit(1)
	}
	exts := []string{"cpp", "cc", "cxx", "h", "hpp", "hh", "h++"}
	dirs := []string{".", "include", "common"}
	for _, dir := range dirs {
		for _, ext := range exts {
			files, _ := filepath.Glob(filepath.Join(dir, "*."+ext))
			for _, f := range files {
				c := exec.Command("clang-format", "-style={BasedOnStyle: Webkit, ColumnLimit: 99}", "-i", f)
				_ = c.Run()
			}
		}
	}
}

func doValgrind(opts BuildOptions) error {
	if err := doBuild(opts); err != nil {
		return err
	}
	exe := executableName()
	if exe == "" {
		return fmt.Errorf("no executable to profile")
	}
	if _, err := exec.LookPath("valgrind"); err != nil {
		return fmt.Errorf("valgrind not found in PATH")
	}
	exePath := dotSlash(exe)
	c := exec.Command("valgrind", "--tool=callgrind", exePath)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: valgrind exited with: %v\n", err)
	}
	callgrindFiles, _ := filepath.Glob("callgrind.out.*")
	if len(callgrindFiles) > 0 && hasCommand("gprof2dot") && hasCommand("dot") {
		c = exec.Command("sh", "-c",
			"gprof2dot -f callgrind "+callgrindFiles[0]+" | dot -Tsvg -o output.svg")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		_ = c.Run()
	}
	if len(callgrindFiles) > 0 && hasCommand("kcachegrind") {
		c = exec.Command("kcachegrind", callgrindFiles[0])
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		_ = c.Run()
	}
	return nil
}

// doDebug builds in debug mode and then launches a debugger.
func doDebug(opts BuildOptions, useClang bool) error {
	if err := doBuild(opts); err != nil {
		return err
	}
	exe := executableName()
	if exe == "" {
		return fmt.Errorf("no executable to debug")
	}
	exePath := dotSlash(exe)

	// Choose debugger
	var debugger string
	if useClang {
		for _, d := range []string{"lldb", "gdb", "cgdb"} {
			if hasCommand(d) {
				debugger = d
				break
			}
		}
	} else {
		for _, d := range []string{"cgdb", "gdb", "lldb"} {
			if hasCommand(d) {
				debugger = d
				break
			}
		}
	}
	if debugger == "" {
		return fmt.Errorf("no debugger found (tried cgdb, gdb, lldb)")
	}

	c := exec.Command(debugger, exePath)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	return c.Run()
}

// doTiny builds a tiny executable and runs sstrip/upx post-processing.
func doTiny(opts BuildOptions) error {
	if err := doBuild(opts); err != nil {
		return err
	}
	exe := executableName()
	if exe == "" {
		return nil
	}
	if opts.Win64 {
		exe += ".exe"
	}
	exePath := dotSlash(exe)

	// Try sstrip
	if hasCommand("sstrip") {
		c := exec.Command("sstrip", exePath)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err == nil {
			fmt.Println("sstrip", exePath)
		}
	}

	// Try upx --brute
	if hasCommand("upx") {
		c := exec.Command("upx", "--brute", exePath)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err == nil {
			fmt.Println("upx --brute", exePath)
		}
	}
	return nil
}

// dotSlash prepends ./ to a relative path to make it executable.
func dotSlash(name string) string {
	if filepath.IsAbs(name) || strings.HasPrefix(name, "."+string(os.PathSeparator)) || strings.HasPrefix(name, "./") {
		return name
	}
	return "." + string(os.PathSeparator) + name
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
