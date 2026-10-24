// Package ui renders the interactive wizard that collects the user's
// codebase options, generates the codebase, and shows the tailored
// getting-started instructions.
package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/VJ-2303/gocraft/internal/generator"
	"github.com/VJ-2303/gocraft/internal/spec"
)

// Run starts the interactive wizard. It blocks until the user completes,
// cancels or aborts the flow, and returns any terminal error.
func Run(ctx context.Context, logger *slog.Logger) error {
	generate := func(rsp spec.ResolvedSpec, target string) (generator.Result, error) {
		return generator.Generate(rsp, absoluteTarget(target), logger, generator.Options{Tidy: true})
	}

	final, err := tea.NewProgram(newWizardModel(generate)).Run()
	if err != nil {
		return err
	}

	model, ok := final.(wizardModel)
	if !ok {
		return fmt.Errorf("unexpected wizard state")
	}

	fmt.Println()

	switch model.outcome {
	case outcomeDone:
		printDone(model.rsp, model.result)
	case outcomeFailed:
		printFailed(model.genErr)
	case outcomeCanceled:
		printCanceled(model.v.targetDir)
	default: // outcomeAborted
		printAborted()
	}

	return nil
}

// absoluteTarget resolves a relative target directory against the
// working directory.
func absoluteTarget(dir string) string {
	if filepath.IsAbs(dir) {
		return dir
	}
	cwd, err := os.Getwd()
	if err != nil {
		return dir
	}
	return filepath.Join(cwd, dir)
}
