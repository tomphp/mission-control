package session

import "time"

// ID uniquely identifies a Claude Code session; it is stable for the
// lifetime of one conversation (the hook payload's session_id).
type ID string

// Session is the tracked state of one running Claude Code session.
type Session struct {
	ID          ID        `json:"id"`
	Cwd         string    `json:"cwd"`
	Folder      string    `json:"folder"`
	GitBranch   string    `json:"gitBranch"`
	State       State     `json:"state"`
	StartedAt   time.Time `json:"startedAt"`
	LastEventAt time.Time `json:"lastEventAt"`
	LastEvent   string    `json:"lastEvent"`
}
