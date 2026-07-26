// Command mission-control is a single binary combining the Claude Code
// hook reporter, the session dashboard webserver, and settings.json
// install/uninstall management.
package main

import (
	"os"

	"github.com/tomoram/mission-control/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
