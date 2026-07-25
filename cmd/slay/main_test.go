package main

import (
	"os"
	"slices"
	"testing"
)

func TestExtractAssignments(t *testing.T) {
	t.Setenv("CXXFLAGS", "")
	t.Setenv("PREFIX", "")

	got := extractAssignments([]string{"strict=1", "debug=0", "PREFIX=/usr", "CXXFLAGS=-O0 -g", "std=c++17", "install"})
	want := []string{"strict", "install"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if prefix := os.Getenv("PREFIX"); prefix != "/usr" {
		t.Errorf("PREFIX = %q, want /usr", prefix)
	}
	if cxxflags := os.Getenv("CXXFLAGS"); cxxflags != "-O0 -g -std=c++17" {
		t.Errorf("CXXFLAGS = %q, want %q", cxxflags, "-O0 -g -std=c++17")
	}
}

// Arguments for the built executable must be passed on untouched
func TestExtractAssignments_RunArgs(t *testing.T) {
	got := extractAssignments([]string{"opt=1", "run", "key=value", "-x"})
	want := []string{"opt", "run", "key=value", "-x"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
