package session

// State is a session's coarse activity state, as shown in the UI.
type State string

const (
	StateWorking         State = "working"
	StateWaitingForInput State = "waiting_for_input"
	StateIdle            State = "idle"
)
