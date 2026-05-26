package main

import (
	"fmt"
	"os"

	. "github.com/xyproto/slay"
)

const versionString = "slay 1.2.0"

func printHelp() {
	fmt.Printf(`%s

slay              - build the project
slay run          - build and run
slay debug        - debug build and launch debugger (gdb/cgdb)
slay debugbuild   - debug build (without launching debugger)
slay debugnosan   - debug build (without sanitizers)
slay opt          - optimized build
slay strict       - build with strict warning flags
slay sloppy       - build with sloppy flags
slay small        - build a smaller executable
slay tiny         - build a tiny executable (+ sstrip/upx)
slay clang        - build using clang++
slay clangdebug   - debug build using clang++ (launches lldb)
slay clangstrict  - use clang++ and strict flags
slay clangsloppy  - use clang++ and sloppy flags
slay clangrebuild - clean and build with clang++
slay clangtest    - build and run tests with clang++
slay clean        - clean all build artifacts and build directories
slay fastclean    - only remove executable and *.o
slay rebuild      - clean and build
slay test         - build and run tests
slay testbuild    - build tests (without running)
slay rec          - profile-guided optimization (build, run, rebuild)
slay fmt          - format source code with clang-format
slay generate     - generate CMakeLists.txt
slay makefile     - generate a standalone Makefile
slay cmake        - build with cmake (prefers ninja, falls back to make)
slay make         - build with make (falls back to cmake+make)
slay ninja        - build with ninja (falls back to cmake+ninja)
slay ninja_install- install from ninja build
slay ninja_clean  - clean ninja build
slay make_install - install from cmake+make build
slay make_clean   - clean cmake+make build
slay pro          - generate QtCreator project file
slay install      - install the project (PREFIX, DESTDIR)
slay pkg          - package the project into pkg/
slay export       - export a standalone Makefile and build.sh
slay script       - generate build.sh and clean.sh
slay valgrind     - build and profile with valgrind
slay win64        - cross-compile for 64-bit Windows
slay smallwin64   - small win64 build
slay tinywin64    - tiny win64 build
slay zap          - build using zapcc++
slay version      - show version
slay -C <dir> ... - run in the given directory
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
		exitOnErr(NewConfig().Build())
	case "rebuild":
		exitOnErr(NewConfig().Rebuild())
	case "clean":
		NewConfig().Clean()
	case "fastclean":
		NewConfig().FastClean()
	case "run":
		exitOnErr(NewConfig().Run(subArgs...))
	case "debug":
		exitOnErr(DebugConfig().LaunchDebugger())
	case "debugbuild":
		exitOnErr(DebugConfig().Build())
	case "debugnosan":
		exitOnErr(DebugNoSanConfig().Build())
	case "opt":
		exitOnErr(OptConfig().Build())
	case "strict":
		exitOnErr(StrictConfig().Build())
	case "sloppy":
		exitOnErr(SloppyConfig().Build())
	case "small":
		exitOnErr(SmallConfig().Build())
	case "tiny":
		exitOnErr(TinyConfig().TinyBuild())
	case "clang":
		exitOnErr(ClangConfig().Build())
	case "clangdebug":
		cfg := ClangConfig()
		cfg.Debug = true
		exitOnErr(cfg.LaunchDebugger())
	case "clangstrict":
		cfg := ClangConfig()
		cfg.Strict = true
		exitOnErr(cfg.Build())
	case "clangsloppy":
		cfg := ClangConfig()
		cfg.Sloppy = true
		exitOnErr(cfg.Build())
	case "clangrebuild":
		exitOnErr(ClangConfig().Rebuild())
	case "clangtest":
		exitOnErr(ClangConfig().Test())
	case "test":
		exitOnErr(NewConfig().Test())
	case "testbuild":
		exitOnErr(NewConfig().TestBuild())
	case "rec":
		exitOnErr(NewConfig().Rec(subArgs...))
	case "fmt":
		exitOnErr(NewConfig().Fmt())
	case "cmake":
		exitOnErr(NewConfig().CMakeBuild())
	case "generate", "cmakelists", "cmakelist", "cmakelists.txt", "CMakeLists.txt":
		exitOnErr(NewConfig().Generate())
	case "pro":
		exitOnErr(NewConfig().Pro())
	case "ninja":
		exitOnErr(NewConfig().Ninja())
	case "ninja_install":
		exitOnErr(NewConfig().NinjaInstall())
	case "ninja_clean":
		NewConfig().NinjaClean()
	case "make_install":
		exitOnErr(NewConfig().CMakeMakeInstall())
	case "make_clean":
		NewConfig().CMakeMakeClean()
	case "install":
		exitOnErr(NewConfig().Install())
	case "pkg":
		exitOnErr(NewConfig().Pkg())
	case "export":
		exitOnErr(NewConfig().Export())
	case "make":
		exitOnErr(NewConfig().Make())
	case "makefile":
		exitOnErr(NewConfig().GenerateMakefile())
	case "script":
		exitOnErr(NewConfig().Script())
	case "valgrind":
		exitOnErr(NewConfig().Valgrind())
	case "win", "win64":
		exitOnErr(Win64Config().Build())
	case "smallwin", "smallwin64":
		exitOnErr(SmallWin64Config().Build())
	case "tinywin", "tinywin64":
		exitOnErr(TinyWin64Config().TinyBuild())
	case "zap":
		exitOnErr(ZapConfig().Build())
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}
