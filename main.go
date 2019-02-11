package main

import (
	"github.com/camptocamp/bivac/cmd"
	_ "github.com/camptocamp/bivac/cmd/all"
)

var exitCode int

// Following variables are filled in by the build script
var version = "<<< filled in by build >>>"

func main() {
	cmd.Execute(version)
}
