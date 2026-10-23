package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/module"

	"github.com/charmbracelet/huh"

	"github.com/vj/nomoreboiler/internal/spec"
)

// askAppType shows the step 1 form: the type of application to generate.
func askAppType(s *spec.ProjectSpec) error {
	appType := string(s.AppType)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Step 1 · What type of application do you want to generate?").
				Description("The application type determines the structure of the generated codebase.").
				Options(
					huh.NewOption("Traditional web application — server-side HTML rendering", string(spec.AppWeb)),
					huh.NewOption("JSON API — processes JSON requests and returns JSON responses", string(spec.AppAPI)),
				).
				Value(&appType),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	s.AppType = spec.AppType(appType)
	return nil
}

// askModulePath shows the module path form. The path is validated live
// with the same rules Go applies to module paths.
func askModulePath(s *spec.ProjectSpec) error {
	modulePath := s.ModulePath

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What is the module path of your project?").
				Description("For example: github.com/alice/example. The final element is used as the project name.").
				Validate(func(value string) error {
					if err := module.CheckPath(value); err != nil {
						return errors.New("not a valid Go module path (e.g. github.com/alice/example)")
					}
					return nil
				}).
				Value(&modulePath),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	s.ModulePath = modulePath
	return nil
}

// collectTargetDir asks for the directory to generate the codebase into.
// The directory must not exist, or must exist and be empty.
func collectTargetDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var target string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Where do you want to generate the codebase?").
				Description("Enter a directory path. It must not exist yet, or must be empty.").
				Validate(func(value string) error {
					if value == "" {
						return errors.New("a directory path is required")
					}
					entries, err := os.ReadDir(value)
					if err != nil {
						if os.IsNotExist(err) {
							return nil
						}
						return fmt.Errorf("cannot use %s: %v", value, err)
					}
					if len(entries) > 0 {
						return errors.New("directory exists and is not empty")
					}
					return nil
				}).
				Value(&target),
		),
	).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return "", err
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(cwd, target)
	}
	return target, nil
}
