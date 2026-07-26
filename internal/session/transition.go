package session

import "github.com/tomoram/mission-control/internal/hook"

// NextState is the single source of truth mapping a hook event to a state
// transition. Events not listed here (including SessionEnd, which is
// handled as a store removal rather than a state transition) are no-ops:
// the current state is returned unchanged.
func NextState(current State, event hook.EventName) (next State, changed bool) {
	switch event {
	case hook.SessionStart:
		return StateIdle, true
	case hook.UserPromptSubmit, hook.PreToolUse, hook.PostToolUse, hook.PostToolUseFailure:
		return StateWorking, true
	case hook.Notification:
		return StateWaitingForInput, true
	case hook.Stop, hook.StopFailure:
		return StateIdle, true
	default:
		return current, false
	}
}
