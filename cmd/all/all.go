package all

import (
	// Run a Bivac agent
	_ "github.com/camptocamp/bivac/cmd/agent"
	// Backup a volume
	_ "github.com/camptocamp/bivac/cmd/backup"
	// Get informations regarding the Bivac manager
	_ "github.com/camptocamp/bivac/cmd/info"
	// Run a Bivac manager
	_ "github.com/camptocamp/bivac/cmd/manager"
	// Run a custom Restic command on a volume's remote repository
	_ "github.com/camptocamp/bivac/cmd/restic"
	// List volumes and display informations regarding the backed up volumes
	_ "github.com/camptocamp/bivac/cmd/volumes"
)
