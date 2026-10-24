// Package cli provides non-interactive, headless execution of gocraft for
// scripts, CI pipelines, and automated agents.
package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/VJ-2303/gocraft/internal/generator"
	"github.com/VJ-2303/gocraft/internal/spec"
)

// Output defines the machine-readable JSON structure returned when --json is passed.
type Output struct {
	Success   bool     `json:"success"`
	TargetDir string   `json:"target_dir,omitempty"`
	Files     []string `json:"files,omitempty"`
	Count     int      `json:"count,omitempty"`
	DryRun    bool     `json:"dry_run,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// Run executes gocraft headlessly with the provided arguments.
func Run(args []string, stdout, stderr io.Writer, logger *slog.Logger) int {
	fs := flag.NewFlagSet("gocraft", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		appType          = fs.String("app", "web", "application type (web|api)")
		modulePath       = fs.String("module", "example.com/myapp", "Go module path")
		database         = fs.String("db", "none", "database dialect (none|sqlite|postgres|mysql)")
		router           = fs.String("router", "servemux", "HTTP router (servemux|flow|chi|gorilla|httprouter)")
		configSource     = fs.String("config", "flags", "configuration source (flags|env)")
		logFormat        = fs.String("log", "plaintext", "log format (plaintext|json|colorized)")
		migrations       = fs.Bool("migrations", false, "include embedded SQL migrations")
		sessions         = fs.Bool("sessions", false, "include server-side session management (web only)")
		requestLogging   = fs.Bool("request-logging", false, "include HTTP request logging middleware")
		basicAuth        = fs.Bool("basic-auth", false, "include HTTP basic auth middleware")
		gitIntegration   = fs.Bool("git-integration", true, "include .gitignore and runtime Git version metadata")
		secureCookies    = fs.Bool("secure-cookies", false, "include signed and encrypted cookie helpers")
		autoHTTPS        = fs.Bool("auto-https", false, "include Let's Encrypt automatic TLS")
		customErrorPages = fs.Bool("custom-error-pages", false, "include custom 400, 401, 404, 500 error pages (web only)")
		emails           = fs.Bool("emails", false, "include SMTP mailer and email templates")
		errorNotifs      = fs.Bool("error-notifications", false, "email runtime errors to an admin address")
		tests            = fs.Bool("tests", true, "generate comprehensive test suites")
		accounts         = fs.String("accounts", "none", "user accounts (none|full|jwt)")
		targetDir        = fs.String("target", "", "destination directory (must be empty or not exist)")
		dryRun           = fs.Bool("dry-run", false, "validate and compute planned files without writing to disk")
		asJSON           = fs.Bool("json", false, "output machine-readable JSON status")
		quiet            = fs.Bool("quiet", false, "suppress progress messages")
		autoDeps         = fs.Bool("auto-deps", true, "automatically enable implied dependencies (e.g. accounts -> sessions, emails)")
	)

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: gocraft [flags] --target <path>\n\nFlags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	fail := func(err error) int {
		if *asJSON {
			_ = json.NewEncoder(stdout).Encode(Output{
				Success: false,
				Error:   err.Error(),
			})
		} else {
			fmt.Fprintf(stderr, "error: %v\n", err)
		}
		return 1
	}

	if *targetDir == "" && !*dryRun {
		return fail(fmt.Errorf("target directory is required: use --target <path>"))
	}

	var absTarget string
	if *targetDir != "" {
		var err error
		absTarget, err = filepath.Abs(*targetDir)
		if err != nil {
			return fail(fmt.Errorf("resolving target directory: %w", err))
		}
	} else {
		absTarget = "/dry-run"
	}

	pSpec := spec.ProjectSpec{
		AppType:            spec.AppType(*appType),
		ModulePath:         *modulePath,
		Database:           spec.Database(*database),
		Migrations:         *migrations,
		Router:             spec.Router(*router),
		ConfigSource:       spec.ConfigSource(*configSource),
		LogFormat:          spec.LogFormat(*logFormat),
		RequestLogging:     *requestLogging,
		BasicAuth:          *basicAuth,
		GitIntegration:     *gitIntegration,
		SecureCookies:      *secureCookies,
		Sessions:           *sessions,
		AutoHTTPS:          *autoHTTPS,
		CustomErrorPages:   *customErrorPages,
		Emails:             *emails,
		ErrorNotifications: *errorNotifs,
		Tests:              *tests,
	}

	switch *accounts {
	case "full":
		pSpec.AccountsFull = true
	case "jwt":
		pSpec.AccountsJWT = true
	case "none", "":
	default:
		return fail(fmt.Errorf("accounts must be 'none', 'full' or 'jwt', got %q", *accounts))
	}

	if *autoDeps {
		pSpec.ApplyDependencies()
	}

	if err := pSpec.Validate(); err != nil {
		return fail(fmt.Errorf("validation error:\n%w", err))
	}

	rsp, err := spec.Resolve(pSpec)
	if err != nil {
		return fail(fmt.Errorf("resolving project spec: %w", err))
	}

	if *dryRun {
		files := generator.PlannedFiles(rsp)
		if *asJSON {
			_ = json.NewEncoder(stdout).Encode(Output{
				Success:   true,
				DryRun:    true,
				TargetDir: absTarget,
				Files:     files,
				Count:     len(files),
			})
		} else if !*quiet {
			fmt.Fprintf(stdout, "gocraft dry-run: %d files would be generated for %s into %s\n", len(files), rsp.ModulePath, absTarget)
			for _, f := range files {
				fmt.Fprintf(stdout, "  %s\n", f)
			}
		}
		return 0
	}

	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	result, err := generator.Generate(rsp, absTarget, logger, generator.Options{Tidy: true})
	if err != nil {
		return fail(fmt.Errorf("generating codebase: %w", err))
	}

	if *asJSON {
		_ = json.NewEncoder(stdout).Encode(Output{
			Success:   true,
			TargetDir: result.Dir,
			Files:     result.Files,
			Count:     len(result.Files),
		})
	} else if !*quiet {
		fmt.Fprintf(stdout, "Successfully generated %s (%d files) in %s\n", rsp.ModulePath, len(result.Files), result.Dir)
	}

	return 0
}
