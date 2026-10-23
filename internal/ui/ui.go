// Package ui renders the interactive wizard that collects the user's
// codebase options, kicks off generation, and shows the tailored
// getting-started instructions.
package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/huh"

	"github.com/vj/nomoreboiler/internal/generator"
	"github.com/vj/nomoreboiler/internal/spec"
)

// Run starts the interactive wizard. It blocks until the user completes or
// cancels the flow, and returns any terminal error.
func Run(ctx context.Context, logger *slog.Logger, version string) error {
	fmt.Printf("nomoreboiler %s — generate a Go web application or JSON API\n\n", version)

	var s spec.ProjectSpec

	for {
		collected, err := collectSpec()
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("Aborted. No codebase was generated.")
				return nil
			}
			return err
		}
		s = collected

		// Auto-enable any features that the selected features depend on,
		// then re-check legality.
		s.ApplyDependencies()

		proceed, err := review(&s)
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("Aborted. No codebase was generated.")
				return nil
			}
			return err
		}
		if proceed {
			break
		}
		// The user chose "start over": loop back to the first form.
	}

	if err := s.Validate(); err != nil {
		return err
	}

	target, err := collectTargetDir()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println("Aborted. No codebase was generated.")
			return nil
		}
		return err
	}

	rsp, err := spec.Resolve(s)
	if err != nil {
		return err
	}

	result, err := runWithSpinner(ctx, func() (generator.Result, error) {
		return generator.Generate(rsp, target, logger, generator.Options{Tidy: true})
	})
	if err != nil {
		return err
	}

	printDone(rsp, result)
	return nil
}

// collectSpec runs the wizard forms and returns the user's answers.
func collectSpec() (spec.ProjectSpec, error) {
	s := spec.Defaults()

	if err := askAppType(&s); err != nil {
		return s, err
	}
	if err := askModulePath(&s); err != nil {
		return s, err
	}
	if err := askCustomize(&s); err != nil {
		return s, err
	}
	if err := askPlusFeatures(&s); err != nil {
		return s, err
	}

	return s, nil
}
