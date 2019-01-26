package agent

import (
	"fmt"
	"os"

	"github.com/camptocamp/bivac/internal/engines"
	"github.com/camptocamp/bivac/internal/utils"
)

func Start() {
	e := &engines.ResticEngine{
		DefaultArgs: []string{
			"--no-cache",
			"--json",
			"-r",
			os.Getenv("RESTIC_REPOSITORY"),
		},
		Output: make(map[string]utils.OutputFormat),
	}

	output := e.Backup()
	fmt.Println(output)
	return
}
