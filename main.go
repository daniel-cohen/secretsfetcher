package main

import "github.com/daniel-cohen/secretsfetcher/cmd"

var (
	// Version is set during build
	Version string = "Unknown"
)

func main() {
	cmd.Execute(Version)
}
