# Orchideous

A Go port of [xyproto/cxx](https://github.com/xyproto/cxx) — an auto-build tool for C and C++ projects. LLMs was used when porting.

## Usage

Place `orchideous` in your `PATH`, then run it in a directory with C/C++ source files:

```sh
orchideous              # build the project
orchideous run          # build and run
orchideous test         # build and run tests
orchideous clean        # remove built files
orchideous opt          # optimized build
orchideous debug        # debug build (with sanitizers)
orchideous clang        # build with clang++
orchideous cmake        # generate CMakeLists.txt
orchideous install      # install to PREFIX
orchideous --help       # show all commands
```

## Installing

(dev version, requires Go 1.26 or later)

```sh
go install github.com/xyproto/orchideous@latest
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
