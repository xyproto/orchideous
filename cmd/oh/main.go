package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xyproto/files"
	"github.com/xyproto/orchideous"
)

const versionString = "oh 1.0.2"

func printHelp() {
	fmt.Printf(`%s

oh              - build the project
oh run          - build and run
oh debug        - debug build and launch debugger (gdb/cgdb)
oh debugbuild   - debug build (without launching debugger)
oh debugnosan   - debug build (without sanitizers)
oh opt          - optimized build
oh strict       - build with strict warning flags
oh sloppy       - build with sloppy flags
oh small        - build a smaller executable
oh tiny         - build a tiny executable (+ sstrip/upx)
oh clang        - build using clang++
oh clangdebug   - debug build using clang++ (launches lldb)
oh clangstrict  - use clang++ and strict flags
oh clangsloppy  - use clang++ and sloppy flags
oh clangrebuild - clean and build with clang++
oh clangtest    - build and run tests with clang++
oh clean        - remove built files
oh fastclean    - only remove executable and *.o
oh rebuild      - clean and build
oh test         - build and run tests
oh testbuild    - build tests (without running)
oh rec          - profile-guided optimization (build, run, rebuild)
oh fmt          - format source code with clang-format
oh cmake        - generate CMakeLists.txt
oh cmake ninja  - generate CMakeLists.txt and build with ninja
oh ninja        - build using existing CMakeLists.txt and ninja
oh ninja_install- install from ninja build
oh ninja_clean  - clean ninja build
oh pro          - generate QtCreator project file
oh install      - install the project (PREFIX, DESTDIR)
oh pkg          - package the project into pkg/
oh export       - export a standalone Makefile and build.sh
oh make         - generate a standalone Makefile
oh script       - generate build.sh and clean.sh
oh valgrind     - build and profile with valgrind
oh win64        - cross-compile for 64-bit Windows
oh smallwin64   - small win64 build
oh tinywin64    - tiny win64 build
oh zap          - build using zapcc++
oh version      - show version
oh -C <dir> ... - run in the given directory
`, versionString)
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func hasCommand(name string) bool {
	return files.WhichCached(name) != ""
}

func doRun(opts orchideous.BuildOptions, runArgs []string) error {
	if err := orchideous.DoBuild(opts); err != nil {
		return err
	}
	exe := orchideous.ExecutableName()
	if exe == "" {
		return fmt.Errorf("no main source file found")
	}
	if opts.Win64 {
		exe += ".exe"
	}
	exePath := orchideous.DotSlash(exe)
	// Auto-detect .exe and use wine if available
	if strings.HasSuffix(exePath, ".exe") {
		if winePath := files.WhichCached("wine"); winePath != "" {
			c := exec.Command(winePath, exePath)
			c.Args = append(c.Args, runArgs...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		}
	}
	c := exec.Command(exePath, runArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func doClean() {
	exe := orchideous.ExecutableName()
	patterns := []string{"*.o", "*.d", "common/*.o", "common/*.d", "include/*.o", "include/*.d", "*.profraw", "*.gcda", "*.gcno", ".sconsign.dblite", "callgrind.out.*"}
	for _, pat := range patterns {
		matches, _ := filepath.Glob(pat)
		for _, f := range matches {
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
	testSrcs := orchideous.GetTestSources()
	for _, ts := range testSrcs {
		testExe := strings.TrimSuffix(ts, filepath.Ext(ts))
		if err := os.Remove(testExe); err == nil {
			fmt.Println("Removed", testExe)
		}
	}
}

func doFastClean() {
	exe := orchideous.ExecutableName()
	matches, _ := filepath.Glob("*.o")
	for _, f := range matches {
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

func doTest(opts orchideous.BuildOptions) error {
	testSrcs := orchideous.GetTestSources()
	if len(testSrcs) == 0 {
		fmt.Println("Nothing to test")
		return nil
	}

	proj := orchideous.DetectProject()
	flags := orchideous.AssembleFlags(proj, opts)
	for _, ts := range testSrcs {
		testExe := strings.TrimSuffix(ts, filepath.Ext(ts))
		srcs := append([]string{ts}, proj.DepSources...)
		if err := orchideous.CompileSources(srcs, testExe, flags); err != nil {
			return fmt.Errorf("building test %s: %w", testExe, err)
		}
		fmt.Printf("Running %s...\n", testExe)
		c := exec.Command(orchideous.DotSlash(testExe))
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("test %s failed: %w", testExe, err)
		}
	}
	return nil
}

func doTestBuild(opts orchideous.BuildOptions) error {
	proj := orchideous.DetectProject()
	flags := orchideous.AssembleFlags(proj, opts)

	if proj.MainSource != "" {
		exe := orchideous.ExecutableName()
		srcs := append([]string{proj.MainSource}, proj.DepSources...)
		if err := orchideous.CompileSources(srcs, exe, flags); err != nil {
			return err
		}
	}

	for _, ts := range proj.TestSources {
		testExe := strings.TrimSuffix(ts, filepath.Ext(ts))
		srcs := append([]string{ts}, proj.DepSources...)
		if err := orchideous.CompileSources(srcs, testExe, flags); err != nil {
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
	if err := orchideous.DoBuild(orchideous.BuildOptions{Opt: true, ProfileGenerate: true}); err != nil {
		return fmt.Errorf("profile generation build: %w", err)
	}
	exe := orchideous.ExecutableName()
	if exe == "" {
		return fmt.Errorf("no executable to run for profiling")
	}
	exePath := orchideous.DotSlash(exe)
	c := exec.Command(exePath, runArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	_ = c.Run()
	return orchideous.DoBuild(orchideous.BuildOptions{Opt: true, ProfileUse: true})
}

func doFmt() {
	if files.WhichCached("clang-format") == "" {
		fmt.Fprintln(os.Stderr, "error: clang-format not found in PATH")
		os.Exit(1)
	}
	exts := []string{"cpp", "cc", "cxx", "h", "hpp", "hh", "h++"}
	dirs := []string{".", "include", "common"}
	for _, dir := range dirs {
		for _, ext := range exts {
			matches, _ := filepath.Glob(filepath.Join(dir, "*."+ext))
			for _, f := range matches {
				c := exec.Command("clang-format", "-style={BasedOnStyle: Webkit, ColumnLimit: 99}", "-i", f)
				_ = c.Run()
			}
		}
	}
}

func doValgrind(opts orchideous.BuildOptions) error {
	if err := orchideous.DoBuild(opts); err != nil {
		return err
	}
	exe := orchideous.ExecutableName()
	if exe == "" {
		return fmt.Errorf("no executable to profile")
	}
	if files.WhichCached("valgrind") == "" {
		return fmt.Errorf("valgrind not found in PATH")
	}
	exePath := orchideous.DotSlash(exe)
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

func doDebug(opts orchideous.BuildOptions, useClang bool) error {
	if err := orchideous.DoBuild(opts); err != nil {
		return err
	}
	exe := orchideous.ExecutableName()
	if exe == "" {
		return fmt.Errorf("no executable to debug")
	}
	exePath := orchideous.DotSlash(exe)

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

func doTiny(opts orchideous.BuildOptions) error {
	if err := orchideous.DoBuild(opts); err != nil {
		return err
	}
	exe := orchideous.ExecutableName()
	if exe == "" {
		return nil
	}
	if opts.Win64 {
		exe += ".exe"
	}
	exePath := orchideous.DotSlash(exe)

	if hasCommand("sstrip") {
		c := exec.Command("sstrip", exePath)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err == nil {
			fmt.Println("sstrip", exePath)
		}
	}

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

func main() {
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
		fmt.Println(versionString)
	case "build":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{}))
	case "rebuild":
		doClean()
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{}))
	case "clean":
		doClean()
	case "fastclean":
		doFastClean()
	case "run":
		exitOnErr(doRun(orchideous.BuildOptions{}, subArgs))
	case "debug":
		exitOnErr(doDebug(orchideous.BuildOptions{Debug: true}, false))
	case "debugbuild":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Debug: true}))
	case "debugnosan":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Debug: true, NoSanitizers: true}))
	case "opt":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Opt: true}))
	case "strict":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Strict: true}))
	case "sloppy":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Sloppy: true}))
	case "small":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Small: true}))
	case "tiny":
		exitOnErr(doTiny(orchideous.BuildOptions{Small: true, Tiny: true}))
	case "clang":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Clang: true}))
	case "clangdebug":
		exitOnErr(doDebug(orchideous.BuildOptions{Clang: true, Debug: true}, true))
	case "clangstrict":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Clang: true, Strict: true}))
	case "clangsloppy":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Clang: true, Sloppy: true}))
	case "clangrebuild":
		doClean()
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Clang: true}))
	case "clangtest":
		exitOnErr(doTest(orchideous.BuildOptions{Clang: true}))
	case "test":
		exitOnErr(doTest(orchideous.BuildOptions{}))
	case "testbuild":
		exitOnErr(doTestBuild(orchideous.BuildOptions{}))
	case "rec":
		exitOnErr(doRec(subArgs))
	case "fmt":
		doFmt()
	case "cmake":
		if len(subArgs) > 0 && subArgs[0] == "ninja" {
			exitOnErr(orchideous.DoCMake(orchideous.BuildOptions{}))
			exitOnErr(orchideous.DoNinja())
		} else {
			exitOnErr(orchideous.DoCMake(orchideous.BuildOptions{}))
		}
	case "pro":
		exitOnErr(orchideous.DoPro(orchideous.BuildOptions{}))
	case "ninja":
		exitOnErr(orchideous.DoNinja())
	case "ninja_install":
		exitOnErr(orchideous.DoNinjaInstall())
	case "ninja_clean":
		orchideous.DoNinjaClean()
	case "install":
		exitOnErr(orchideous.DoInstall())
	case "pkg":
		exitOnErr(orchideous.DoPkg())
	case "export":
		exitOnErr(orchideous.DoExport())
	case "make":
		exitOnErr(orchideous.DoMakeFile())
	case "script":
		exitOnErr(orchideous.DoScript())
	case "valgrind":
		exitOnErr(doValgrind(orchideous.BuildOptions{}))
	case "win", "win64":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Win64: true}))
	case "smallwin", "smallwin64":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Win64: true, Small: true}))
	case "tinywin", "tinywin64":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Win64: true, Small: true, Tiny: true}))
	case "zap":
		exitOnErr(orchideous.DoBuild(orchideous.BuildOptions{Zap: true}))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}
