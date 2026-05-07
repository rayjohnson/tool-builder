package main

import "github.com/rayjohnson/tool-builder/cmd"

// Set by GoReleaser via -ldflags at build time.
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	cmd.Execute(version, buildTime)
}
