package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vj/nomoreboiler/internal/ui"
)

// version holds the version of the application. It is overridden at build
// time using -ldflags "-X main.version=..." when release binaries are built.
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run performs the startup and shutdown of the application. All errors are
// returned and handled by main, keeping this function's control flow simple.
func run() error {
	// Parse the configuration settings for the application from the
	// command-line flags.
	var cfg config
	var showVersion bool

	flag.BoolVar(&showVersion, "version", false, "display version and exit")
	flag.StringVar(&cfg.logFormat, "log-format", "plaintext", "log format (plaintext|json)")
	flag.IntVar(&cfg.logLevel, "log-level", int(slog.LevelInfo), "minimum log level (debug=-4, info=0, warn=4, error=8)")
	flag.Parse()

	if showVersion {
		fmt.Printf("nomoreboiler %s\n", version)
		return nil
	}

	logger := newLogger(cfg)

	// Initialize a context that is canceled when the user interrupts the
	// application, so that any in-flight generation work stops cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return ui.Run(ctx, logger, version)
}

// config holds the configuration settings for the application.
type config struct {
	logFormat string
	logLevel  int
}

// newLogger returns a slog.Logger writing to standard error using the
// configured format. The tool's own logging is deliberately plain: the
// interesting output is what the UI renders on standard output.
func newLogger(cfg config) *slog.Logger {
	var handler slog.Handler

	level := slog.Level(cfg.logLevel)
	opts := &slog.HandlerOptions{Level: level}

	switch cfg.logFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	return slog.New(handler)
}
