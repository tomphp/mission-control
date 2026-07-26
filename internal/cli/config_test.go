package cli

import "testing"

func TestPortDefault(t *testing.T) {
	t.Setenv("MISSION_CONTROL_PORT", "")
	if got := port(); got != defaultPort {
		t.Errorf("port() = %d, want %d", got, defaultPort)
	}
}

func TestPortFromEnv(t *testing.T) {
	t.Setenv("MISSION_CONTROL_PORT", "9999")
	if got := port(); got != 9999 {
		t.Errorf("port() = %d, want 9999", got)
	}
}

func TestPortInvalidEnvFallsBackToDefault(t *testing.T) {
	t.Setenv("MISSION_CONTROL_PORT", "not-a-number")
	if got := port(); got != defaultPort {
		t.Errorf("port() = %d, want %d", got, defaultPort)
	}
}

func TestPortNonPositiveEnvFallsBackToDefault(t *testing.T) {
	t.Setenv("MISSION_CONTROL_PORT", "-1")
	if got := port(); got != defaultPort {
		t.Errorf("port() = %d, want %d", got, defaultPort)
	}
}

func TestBaseURL(t *testing.T) {
	if got, want := baseURL(4317), "http://127.0.0.1:4317"; got != want {
		t.Errorf("baseURL(4317) = %q, want %q", got, want)
	}
}
