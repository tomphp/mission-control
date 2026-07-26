package session_test

import (
	"testing"

	"github.com/tomoram/mission-control/internal/hook"
	"github.com/tomoram/mission-control/internal/session"
)

func TestNextState_StateChangingEvents(t *testing.T) {
	allStates := []session.State{session.StateWorking, session.StateWaitingForInput, session.StateIdle}

	tests := []struct {
		event hook.EventName
		want  session.State
	}{
		{hook.SessionStart, session.StateIdle},
		{hook.UserPromptSubmit, session.StateWorking},
		{hook.PreToolUse, session.StateWorking},
		{hook.PostToolUse, session.StateWorking},
		{hook.PostToolUseFailure, session.StateWorking},
		{hook.Notification, session.StateWaitingForInput},
		{hook.Stop, session.StateIdle},
		{hook.StopFailure, session.StateIdle},
	}

	for _, tt := range tests {
		for _, current := range allStates {
			t.Run(string(tt.event)+"/from_"+string(current), func(t *testing.T) {
				got, changed := session.NextState(current, tt.event)
				if got != tt.want {
					t.Errorf("NextState(%s, %s) state = %s, want %s", current, tt.event, got, tt.want)
				}
				if !changed {
					t.Errorf("NextState(%s, %s) changed = false, want true", current, tt.event)
				}
			})
		}
	}
}

func TestNextState_UnknownEventIsNoOp(t *testing.T) {
	allStates := []session.State{session.StateWorking, session.StateWaitingForInput, session.StateIdle}
	unknownEvents := []hook.EventName{
		"PreCompact", "SubagentStop", "SessionEnd", "SomeFutureEvent", "",
	}

	for _, current := range allStates {
		for _, ev := range unknownEvents {
			got, changed := session.NextState(current, ev)
			if got != current {
				t.Errorf("NextState(%s, %s) state = %s, want unchanged %s", current, ev, got, current)
			}
			if changed {
				t.Errorf("NextState(%s, %s) changed = true, want false", current, ev)
			}
		}
	}
}
