// Package cli wires together mission-control's subcommands: the single
// binary is invoked as `report` (by Claude Code hooks), `serve` (the
// webserver), and `install`/`uninstall` (settings.json management).
package cli

import (
	"fmt"
	"os"
)

func Run(args []string) int {
	if len(args) < 1 {
		printUsage()
		return 2
	}

	switch args[0] {
	case "report":
		return runReport(args[1:])
	case "serve":
		return runServe(args[1:])
	case "install":
		return runInstall(args[1:])
	case "uninstall":
		return runUninstall(args[1:])
	case "-h", "--help", "help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "mission-control: unknown subcommand %q\n\n", args[0])
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `mission-control — a live dashboard of running Claude Code sessions

Usage:
  mission-control serve [--port N]                   Run the webserver + dashboard
  mission-control install [--project] [--dry-run]     Register hooks in Claude Code settings.json
  mission-control uninstall [--project] [--dry-run]   Remove mission-control's hooks
  mission-control report                              (internal) invoked by Claude Code hooks`)
}
