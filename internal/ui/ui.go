// Package ui renders the interactive wizard that collects the user's
// codebase options and kicks off generation.
package ui

import (
	"context"
	"fmt"
	"log/slog"
)

// Run starts the interactive wizard. It blocks until the user completes or
// cancels the flow, and returns any terminal error.
func Run(ctx context.Context, logger *slog.Logger, version string) error {
	// Placeholder for the skeleton milestone: the Bubble Tea program lands
	// with the generator engine.
	fmt.Println("nomoreboiler", version, "— wizard under construction")
	logger.DebugContext(ctx, "ui started")
	return nil
}
