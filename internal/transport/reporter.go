// Package transport is how the CLI's "report" subcommand delivers an
// enriched hook event to the mission-control server.
package transport

import (
	"context"

	"github.com/tomoram/mission-control/internal/hook"
)

// Event is a hook payload enriched with information only the CLI can derive
// (currently just the git branch, since the server has no filesystem access
// to the reporting machine in the general case).
type Event struct {
	hook.Payload
	GitBranch string `json:"git_branch"`
}

// Reporter delivers an Event to wherever the server is listening. It is an
// interface so the CLI's report command isn't hard-wired to HTTP.
type Reporter interface {
	Report(ctx context.Context, event Event) error
}
