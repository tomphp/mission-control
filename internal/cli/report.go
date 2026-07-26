package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/tomoram/mission-control/internal/gitinfo"
	"github.com/tomoram/mission-control/internal/hook"
	"github.com/tomoram/mission-control/internal/transport"
)

const reportTimeout = 300 * time.Millisecond

// runReport is what Claude Code's hooks invoke: read the event off stdin,
// enrich it with the git branch, and report it to the server if one is
// listening. It always exits 0 — a hook must never block or fail a Claude
// Code session because the dashboard isn't running.
func runReport(_ []string) int {
	payload, err := hook.Parse(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mission-control report: %v\n", err)
		return 0
	}

	branch, _ := gitinfo.Branch(payload.Cwd)
	event := transport.Event{Payload: payload, GitBranch: branch}

	reporter := transport.NewHTTPReporter(baseURL(port()), reportTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), reportTimeout)
	defer cancel()

	if err := reporter.Report(ctx, event); err != nil {
		fmt.Fprintf(os.Stderr, "mission-control report: %v\n", err)
	}
	return 0
}
