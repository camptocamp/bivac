package main

import (
	"runtime"

	"github.com/camptocamp/bivac/cmd"
	_ "github.com/camptocamp/bivac/cmd/all"
	"github.com/camptocamp/bivac/internal/utils"
)

var (
	exitCode  int
	buildInfo utils.BuildInfo

	// Following variables are filled in by the build script
	version    = "<<< filled in by build >>>"
	buildDate  = "<<< filled in by build >>>"
	commitSha1 = "<<< filled in by build >>>"
)

func main() {
	buildInfo.Version = version
	buildInfo.Date = buildDate
	buildInfo.CommitSha1 = commitSha1
	buildInfo.Runtime = runtime.Version()
	cmd.Execute(buildInfo)
}
