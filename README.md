# Orchideous

A Go port of [xyproto/cxx](https://github.com/xyproto/cxx) — an auto-build tool for C and C++ projects. LLMs was used when porting.

## Usage

Place `orchideous` in your `PATH` as `oh`, then run it in a directory with C/C++ source files:

```sh
oh              # build the project
oh run          # build and run
oh test         # build and run tests
oh clean        # remove built files
oh opt          # optimized build
oh debug        # debug build (with sanitizers)
oh clang        # build with clang++
oh cmake        # generate CMakeLists.txt
oh install      # install to PREFIX
oh --help       # show all commands
```

## Installing

On Arch Linux, using the dev version and Go 1.26 or later:

```sh
go install github.com/xyproto/orchideous@latest
install -Dm755 ~/go/bin/orchideous /usr/bin/oh
```

On other UNIX-like systems:

```sh
go install github.com/xyproto/orchideous@latest
install -m755 ~/go/bin/orchideous /usr/local/bin/oh
```


## Building

```sh
cd orchideous
go build -o orchideous .
```

## Features

- Auto-detects main source, dependencies, test files, and include paths
- Incremental compilation (only recompiles changed sources)
- pkg-config integration for external library discovery
- Profile-guided optimization (`orchideous rec`)
- CMake, QtCreator, Makefile, and build script generation
- Win64 cross-compilation support
- Qt6, Boost, OpenGL, SDL2, Vulkan, GTK, and more

## General

* License: BSD-3
* Version: 0.0.1
