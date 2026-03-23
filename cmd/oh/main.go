package main

import (
	"fmt"
	"os"

	"github.com/xyproto/orchideous"
)

const versionString = "oh 1.0.6"

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
		exitOnErr(orchideous.NewConfig().Build())
	case "rebuild":
		exitOnErr(orchideous.NewConfig().Rebuild())
	case "clean":
		orchideous.NewConfig().Clean()
	case "fastclean":
		orchideous.NewConfig().FastClean()
	case "run":
		exitOnErr(orchideous.NewConfig().Run(subArgs...))
	case "debug":
		exitOnErr(orchideous.DebugConfig().LaunchDebugger())
	case "debugbuild":
		exitOnErr(orchideous.DebugConfig().Build())
	case "debugnosan":
		exitOnErr(orchideous.DebugNoSanConfig().Build())
	case "opt":
		exitOnErr(orchideous.OptConfig().Build())
	case "strict":
		exitOnErr(orchideous.StrictConfig().Build())
	case "sloppy":
		exitOnErr(orchideous.SloppyConfig().Build())
	case "small":
		exitOnErr(orchideous.SmallConfig().Build())
	case "tiny":
		exitOnErr(orchideous.TinyConfig().TinyBuild())
	case "clang":
		exitOnErr(orchideous.ClangConfig().Build())
	case "clangdebug":
		cfg := orchideous.ClangConfig()
		cfg.Debug = true
		exitOnErr(cfg.LaunchDebugger())
	case "clangstrict":
		cfg := orchideous.ClangConfig()
		cfg.Strict = true
		exitOnErr(cfg.Build())
	case "clangsloppy":
		cfg := orchideous.ClangConfig()
		cfg.Sloppy = true
		exitOnErr(cfg.Build())
	case "clangrebuild":
		exitOnErr(orchideous.ClangConfig().Rebuild())
	case "clangtest":
		exitOnErr(orchideous.ClangConfig().Test())
	case "test":
		exitOnErr(orchideous.NewConfig().Test())
	case "testbuild":
		exitOnErr(orchideous.NewConfig().TestBuild())
	case "rec":
		exitOnErr(orchideous.NewConfig().Rec(subArgs...))
	case "fmt":
		exitOnErr(orchideous.NewConfig().Fmt())
	case "cmake":
		if len(subArgs) > 0 && subArgs[0] == "ninja" {
			exitOnErr(orchideous.NewConfig().CMakeNinja())
		} else {
			exitOnErr(orchideous.NewConfig().CMake())
		}
	case "pro":
		exitOnErr(orchideous.NewConfig().Pro())
	case "ninja":
		exitOnErr(orchideous.NewConfig().Ninja())
	case "ninja_install":
		exitOnErr(orchideous.NewConfig().NinjaInstall())
	case "ninja_clean":
		orchideous.NewConfig().NinjaClean()
	case "install":
		exitOnErr(orchideous.NewConfig().Install())
	case "pkg":
		exitOnErr(orchideous.NewConfig().Pkg())
	case "export":
		exitOnErr(orchideous.NewConfig().Export())
	case "make":
		exitOnErr(orchideous.NewConfig().MakeFile())
	case "script":
		exitOnErr(orchideous.NewConfig().Script())
	case "valgrind":
		exitOnErr(orchideous.NewConfig().Valgrind())
	case "win", "win64":
		exitOnErr(orchideous.Win64Config().Build())
	case "smallwin", "smallwin64":
		exitOnErr(orchideous.SmallWin64Config().Build())
	case "tinywin", "tinywin64":
		exitOnErr(orchideous.TinyWin64Config().TinyBuild())
	case "zap":
		exitOnErr(orchideous.ZapConfig().Build())
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}
