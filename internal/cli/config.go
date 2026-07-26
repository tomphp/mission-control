package cli

import (
	"fmt"
	"os"
	"strconv"
)

const defaultPort = 47923

// port resolves the server port: MISSION_CONTROL_PORT if set, else the
// default. Both `serve` and `report` use this so the registered hook
// command never needs to hard-code a port.
func port() int {
	if v := os.Getenv("MISSION_CONTROL_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			return p
		}
	}
	return defaultPort
}

func baseURL(p int) string {
	return fmt.Sprintf("http://127.0.0.1:%d", p)
}
