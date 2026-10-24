package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/VJ-2303/gocraft/internal/cli"
	"github.com/VJ-2303/gocraft/internal/ui"
	"github.com/VJ-2303/gocraft/internal/update"
	"golang.org/x/term"
)

// version holds the version of the application. It is overridden at build
// time using -ldflags "-X main.version=..." when release binaries are built.
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "upgrade" {
		os.Exit(update.Run(os.Args[2:], os.Stdout, os.Stderr, version))
	}
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run performs the startup and shutdown of the application. All errors are
// returned and handled by main, keeping this function's control flow simple.
func run() error {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-version", "-v":
			fmt.Printf("gocraft %s\n", version)
			return nil
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	isInteractive := term.IsTerminal(int(os.Stdin.Fd()))
	hasFlags := len(os.Args) > 1

	if !suppressUpdateNotice(os.Args[1:]) {
		update.Notice(os.Stderr, version)
	}

	if hasFlags || !isInteractive {
		exitCode := cli.Run(os.Args[1:], os.Stdout, os.Stderr, logger)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return ui.Run(ctx, logger)
}

// suppressUpdateNotice keeps scripted output clean: machine-readable, quiet,
// and help runs never wait for or print the hint.
func suppressUpdateNotice(args []string) bool {
	for _, arg := range args {
		name := strings.TrimLeft(arg, "-")
		if i := strings.IndexByte(name, '='); i >= 0 {
			name = name[:i]
		}
		switch name {
		case "json", "quiet", "help", "h":
			return true
		}
	}
	return false
}
