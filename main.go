/*
transmission-jobs is a job runner that uses Transmission RPC calls to manage torrent inventory.

The jobs module isn't meant to have a stable API unless someone else uses it, so just let @mark-ignacio know via a
Github issue.
*/
package main

import "github.com/mark-ignacio/transmission-jobs/internal/cmd"

//go:generate go run internal/gen/expr.go

func main() {
	cmd.Execute()
}
