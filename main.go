package main

import "github.com/rayjohnson/tool-builder/cmd"

// Set by GoReleaser or Makefile via -ldflags at build time.
var (
	version   = "dev"
	buildTime = "unknown"
	moduleDir = "" // set by Makefile for dev builds; empty for released binaries
)

func main() {
	cmd.Execute(version, buildTime, moduleDir)
}
