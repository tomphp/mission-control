package cli

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tomoram/mission-control/internal/server"
	"github.com/tomoram/mission-control/internal/session"
	webui "github.com/tomoram/mission-control/web"
)

const (
	sessionTTL  = 24 * time.Hour
	pruneEvery  = time.Minute
	shutdownWin = 5 * time.Second
)

func runServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	p := fs.Int("port", port(), "port to listen on (default from MISSION_CONTROL_PORT or 47923)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	uiFS, err := webui.DistFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mission-control serve: %v\n", err)
		return 1
	}

	store := session.NewStore()
	broadcaster := server.NewSSEBroadcaster()
	handler := server.New(store, broadcaster, uiFS)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go server.RunPruneLoop(ctx, store, broadcaster, sessionTTL, pruneEvery)

	addr := fmt.Sprintf("127.0.0.1:%d", *p)
	httpServer := &http.Server{Addr: addr, Handler: handler}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownWin)
		defer cancel()
		httpServer.Shutdown(shutdownCtx)
	}()

	fmt.Fprintf(os.Stderr, "mission-control: listening on http://%s\n", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "mission-control serve: %v\n", err)
		return 1
	}
	return 0
}
