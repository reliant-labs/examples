// yours: scaffolded once, never touched again — forge will not overwrite this file
// forge writes it ONCE (write-if-absent). The command tree + component wiring is
// owned code — add a component by hand-editing here (or via `forge scaffold`), never a
// re-derivation. Deleting it sticks; `forge project rescaffold <path>` re-creates it from template.
// Source: internal/templates/project/cmd-main.go.tmpl.
//
// The COMPOSITION ROOT: all command logic lives under cmd/notes/cmd
// (package cmd — root.go assembles the tree), and cmd.Execute builds it.
// What a process RUNS is a root of the composition graph
// (internal/app/graph_gen.go): `notes server` runs the server root — every
// service, worker and operator in one process. Splitting a process is a new
// root (a `//forge:root` struct in internal/app), and each root's command is
// generated (cmd/notes/cmd/roots_gen.go).
package main

import (
	"example.com/notes/cmd/notes/cmd"
)

func main() {
	cmd.Execute()
}
