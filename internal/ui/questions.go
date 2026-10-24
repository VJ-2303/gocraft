package ui

import (
	"fmt"
	"os"

	"github.com/VJ-2303/gocraft/internal/spec"
	"golang.org/x/mod/module"
)

type questionKind int

const (
	kindSelect questionKind = iota
	kindBool
	kindText
	kindReview
)

type option struct {
	label string
	value string
}

type question struct {
	id          string
	category    string
	title       string
	desc        string
	kind        questionKind
	placeholder string

	options  func(v *wizardValues) []option
	visible  func(v *wizardValues) bool
	str      *string
	boolean  *bool
	validate func(string) error
}

func getQuestions(v *wizardValues) []question {
	return []question{
		{
			id:       "app_type",
			category: "Architecture",
			title:    "Application Type",
			desc:     "Defines codebase structure, HTTP response conventions, and asset pipeline.",
			kind:     kindSelect,
			str:      &v.appType,
			options: func(*wizardValues) []option {
				return []option{
					{"Traditional web application — server-side HTML rendering", string(spec.AppWeb)},
					{"JSON API — structured JSON request and response payloads", string(spec.AppAPI)},
				}
			},
		},
		{
			id:          "module",
			category:    "Architecture",
			title:       "Module Path",
			desc:        "Go module import path used throughout every package declaration.",
			kind:        kindText,
			str:         &v.modulePath,
			placeholder: "example.com/myapp",
			validate: func(val string) error {
				if val == "" {
					return fmt.Errorf("module path is required")
				}
				if err := module.CheckPath(val); err != nil {
					return fmt.Errorf("invalid Go module path: %w", err)
				}
				return nil
			},
		},
		{
			id:       "database",
			category: "Persistence",
			title:    "Database Engine",
			desc:     "SQL database driver, connection pool lifecycle, and query models.",
			kind:     kindSelect,
			str:      &v.database,
			options: func(*wizardValues) []option {
				return []option{
					{"None — stateless application without database connectivity", string(spec.DBNone)},
					{"SQLite — lightweight, CGO-free modernc.org/sqlite embedded database", string(spec.DBSQLite)},
					{"PostgreSQL — production pgx/v5 connection pool with binary protocols", string(spec.DBPostgres)},
					{"MySQL — battle-tested go-sql-driver with native datetime parsing", string(spec.DBMySQL)},
				}
			},
		},
		{
			id:       "router",
			category: "Architecture",
			title:    "HTTP Router",
			desc:     "Request multiplexer powering route declarations and path parameters.",
			kind:     kindSelect,
			str:      &v.router,
			options: func(*wizardValues) []option {
				return []option{
					{"http.ServeMux — Go standard library multiplexer with method matching", string(spec.RouterServeMux)},
					{"Flow — lightweight pattern-matching router by Alex Edwards", string(spec.RouterFlow)},
					{"Chi — composable, 100% standard-library-compatible middleware router", string(spec.RouterChi)},
					{"Gorilla Mux — feature-rich router with regex and prefix support", string(spec.RouterGorilla)},
					{"HttpRouter — high-performance radix-tree routing engine", string(spec.RouterHttpRouter)},
				}
			},
		},
		{
			id:       "config",
			category: "Environment",
			title:    "Configuration Source",
			desc:     "Where application settings and secrets are parsed on startup.",
			kind:     kindSelect,
			str:      &v.configSource,
			options: func(*wizardValues) []option {
				return []option{
					{"Command-line flags — parsed via standard flag package with --help overview", string(spec.ConfigFlags)},
					{"Environment variables — read via internal/env with fail-fast validation", string(spec.ConfigEnv)},
				}
			},
		},
		{
			id:       "log",
			category: "Environment",
			title:    "Log Format",
			desc:     "slog structured logger output format sent to standard output.",
			kind:     kindSelect,
			str:      &v.logFormat,
			options: func(*wizardValues) []option {
				return []option{
					{"Plaintext — standard human-readable key=value output", string(spec.LogPlaintext)},
					{"JSON — machine-readable JSON objects for log aggregators", string(spec.LogJSON)},
					{"Colorized — dev-friendly timestamped logs with colored levels", string(spec.LogColorized)},
				}
			},
		},
		{
			id:       "migrations",
			category: "Persistence",
			title:    "SQL Migrations",
			desc:     "Embedded SQL migrations with startup auto-apply and Makefile targets.",
			kind:     kindBool,
			boolean:  &v.migrations,
			visible: func(v *wizardValues) bool {
				return v.database != string(spec.DBNone)
			},
		},
		{
			id:       "sessions",
			category: "Security",
			title:    "Server-Side Sessions",
			desc:     "Encrypted, database-backed sessions using alexedwards/scs.",
			kind:     kindBool,
			boolean:  &v.sessions,
			visible: func(v *wizardValues) bool {
				return v.appType == string(spec.AppWeb)
			},
		},
		{
			id:       "request_logging",
			category: "Observability",
			title:    "Request Logging",
			desc:     "Middleware recording client IP, method, URI, protocol, status, and size.",
			kind:     kindBool,
			boolean:  &v.requestLogging,
		},
		{
			id:       "basic_auth",
			category: "Security",
			title:    "HTTP Basic Authentication",
			desc:     "bcrypt-hashed password verification protecting endpoints with 401 challenges.",
			kind:     kindBool,
			boolean:  &v.basicAuth,
		},
		{
			id:       "git_integration",
			category: "Environment",
			title:    "Git Integration",
			desc:     "Production .gitignore and runtime Git version & revision metadata helpers.",
			kind:     kindBool,
			boolean:  &v.gitIntegration,
		},
		{
			id:       "secure_cookies",
			category: "Security",
			title:    "Secure Cookies",
			desc:     "HMAC-SHA256 signing and AES-256-GCM encryption helpers in internal/cookies.",
			kind:     kindBool,
			boolean:  &v.secureCookies,
		},
		{
			id:       "auto_https",
			category: "Platform",
			title:    "Automatic HTTPS",
			desc:     "Let's Encrypt TLS certificates via autocert with port 80 HTTP-to-HTTPS redirect.",
			kind:     kindBool,
			boolean:  &v.autoHTTPS,
		},
		{
			id:       "custom_error_pages",
			category: "Presentation",
			title:    "Custom Error Pages",
			desc:     "Branded, friendly 400, 401, 404, and 500 error templates in pages/errors/.",
			kind:     kindBool,
			boolean:  &v.customErrorPages,
			visible: func(v *wizardValues) bool {
				return v.appType == string(spec.AppWeb)
			},
		},
		{
			id:       "emails",
			category: "Services",
			title:    "Outgoing Emails",
			desc:     "SMTP mailer helpers with multi-part text and HTML email templates.",
			kind:     kindBool,
			boolean:  &v.emails,
		},
		{
			id:       "error_notifications",
			category: "Observability",
			title:    "Runtime Error Notifications",
			desc:     "Automatically email panic and 500 error stack traces to an admin address.",
			kind:     kindBool,
			boolean:  &v.errorNotifs,
			visible: func(v *wizardValues) bool {
				return v.emails
			},
		},
		{
			id:       "tests",
			category: "Quality",
			title:    "Test Suites",
			desc:     "Comprehensive handler, route, and authentication test suite with helpers.",
			kind:     kindBool,
			boolean:  &v.tests,
		},
		{
			id:       "accounts",
			category: "Security",
			title:    "User Accounts & Authentication",
			desc:     "Complete identity lifecycle with password hashing, sessions, or JWT tokens.",
			kind:     kindSelect,
			str:      &v.accounts,
			options: func(v *wizardValues) []option {
				opts := []option{{"None — no user authentication system", "none"}}
				if v.appType == string(spec.AppWeb) {
					opts = append(opts, option{"Full — signup, login, logout, password reset, and CSRF protection", "full"})
				} else {
					opts = append(opts, option{"JWT — JSON Web Token registration, activation, and bearer tokens", "jwt"})
				}
				return opts
			},
		},
		{
			id:          "target",
			category:    "Output",
			title:       "Target Directory",
			desc:        "Local file path to scaffold into. Must not exist, or must be empty.",
			kind:        kindText,
			str:         &v.targetDir,
			placeholder: "./my-app",
			validate: func(val string) error {
				if val == "" {
					return fmt.Errorf("target directory is required")
				}
				info, err := os.Stat(val)
				if os.IsNotExist(err) {
					return nil
				}
				if err != nil {
					return fmt.Errorf("inspecting target: %w", err)
				}
				if !info.IsDir() {
					return fmt.Errorf("target %q exists and is a file, not a directory", val)
				}
				entries, err := os.ReadDir(val)
				if err != nil {
					return fmt.Errorf("reading target directory: %w", err)
				}
				if len(entries) > 0 {
					return fmt.Errorf("target directory %q exists and is not empty", val)
				}
				return nil
			},
		},
		{
			id:       "review",
			category: "Generate",
			title:    "Configuration Review",
			desc:     "Verify selections. Press a number key [1-9] to edit any step directly, or Enter to generate.",
			kind:     kindReview,
		},
	}
}
