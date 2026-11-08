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
	id             string
	category       string
	title          string
	desc           string
	kind           questionKind
	placeholder    string
	defaultOnEmpty string

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
			desc:     "Codebase shape: web app with HTML, or a JSON API.",
			kind:     kindSelect,
			str:      &v.appType,
			options: func(*wizardValues) []option {
				return []option{
					{"Web app — server-rendered HTML", string(spec.AppWeb)},
					{"JSON API", string(spec.AppAPI)},
				}
			},
		},
		{
			id:          "module",
			category:    "Architecture",
			title:       "Module Path",
			desc:        "Import path for every generated package.",
			kind:        kindText,
			str:         &v.modulePath,
			placeholder: "example.com/myapp",
			validate: func(val string) error {
				if val == "" {
					return fmt.Errorf("module path is required (e.g. github.com/you/app)")
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
			desc:     "Driver, pooling and query models.",
			kind:     kindSelect,
			str:      &v.database,
			options: func(*wizardValues) []option {
				return []option{
					{"None — no database", string(spec.DBNone)},
					{"SQLite — embedded, no CGO", string(spec.DBSQLite)},
					{"PostgreSQL — pgx/v5 pool", string(spec.DBPostgres)},
					{"MySQL — go-sql-driver", string(spec.DBMySQL)},
				}
			},
		},
		{
			id:       "router",
			category: "Architecture",
			title:    "HTTP Router",
			desc:     "Route declarations and path parameters.",
			kind:     kindSelect,
			str:      &v.router,
			options: func(*wizardValues) []option {
				return []option{
					{"http.ServeMux — standard library, method matching", string(spec.RouterServeMux)},
					{"Flow — tiny pattern matcher", string(spec.RouterFlow)},
					{"Chi — composable, stdlib-compatible", string(spec.RouterChi)},
					{"Gorilla mux — regex and prefixes", string(spec.RouterGorilla)},
					{"httprouter — radix tree, fast", string(spec.RouterHttpRouter)},
				}
			},
		},
		{
			id:       "config",
			category: "Environment",
			title:    "Configuration Source",
			desc:     "Where settings and secrets are read at startup.",
			kind:     kindSelect,
			str:      &v.configSource,
			options: func(*wizardValues) []option {
				return []option{
					{"Command-line flags — standard flag package", string(spec.ConfigFlags)},
					{"Environment variables — via internal/env", string(spec.ConfigEnv)},
				}
			},
		},
		{
			id:       "log",
			category: "Environment",
			title:    "Log Format",
			desc:     "slog output format.",
			kind:     kindSelect,
			str:      &v.logFormat,
			options: func(*wizardValues) []option {
				return []option{
					{"Plaintext — key=value", string(spec.LogPlaintext)},
					{"JSON — for log aggregators", string(spec.LogJSON)},
					{"Colorized — dev, colored levels", string(spec.LogColorized)},
				}
			},
		},
		{
			id:       "migrations",
			category: "Persistence",
			title:    "SQL Migrations",
			desc:     "Embedded SQL, auto-applied on start. Makefile targets included.",
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
			desc:     "Logs IP, method, URI, status and size per request.",
			kind:     kindBool,
			boolean:  &v.requestLogging,
		},
		{
			id:       "basic_auth",
			category: "Security",
			title:    "HTTP Basic Authentication",
			desc:     "Protect endpoints with bcrypt basic auth.",
			kind:     kindBool,
			boolean:  &v.basicAuth,
		},
		{
			id:       "git_integration",
			category: "Environment",
			title:    "Git Integration",
			desc:     ".gitignore plus runtime version metadata from git.",
			kind:     kindBool,
			boolean:  &v.gitIntegration,
		},
		{
			id:       "secure_cookies",
			category: "Security",
			title:    "Secure Cookies",
			desc:     "HMAC-SHA256 / AES-256-GCM helpers in internal/cookies.",
			kind:     kindBool,
			boolean:  &v.secureCookies,
		},
		{
			id:       "auto_https",
			category: "Platform",
			title:    "Automatic HTTPS",
			desc:     "Let's Encrypt TLS via autocert. Port 80 redirects to HTTPS.",
			kind:     kindBool,
			boolean:  &v.autoHTTPS,
		},
		{
			id:       "custom_error_pages",
			category: "Presentation",
			title:    "Custom Error Pages",
			desc:     "Branded 400, 401, 404, 500 pages in pages/errors/.",
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
			desc:     "SMTP mailer with multipart text/HTML templates.",
			kind:     kindBool,
			boolean:  &v.emails,
		},
		{
			id:       "error_notifications",
			category: "Observability",
			title:    "Runtime Error Notifications",
			desc:     "Email panics and 500 stack traces to an admin.",
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
			desc:     "Handler, route and auth tests with helpers.",
			kind:     kindBool,
			boolean:  &v.tests,
		},
		{
			id:       "accounts",
			category: "Security",
			title:    "User Accounts & Authentication",
			desc:     "Signup, login, reset; sessions for web, JWT for APIs.",
			kind:     kindSelect,
			str:      &v.accounts,
			options: func(v *wizardValues) []option {
				opts := []option{{"None", "none"}}
				if v.appType == string(spec.AppWeb) {
					opts = append(opts, option{"Full — signup, login, reset, CSRF", "full"})
				} else {
					opts = append(opts, option{"JWT — register, activate, bearer tokens", "jwt"})
				}
				return opts
			},
		},
		{
			id:             "target",
			category:       "Output",
			title:          "Target Directory",
			desc:           "Must not exist or be empty. Defaults to ./my-app.",
			kind:           kindText,
			str:            &v.targetDir,
			placeholder:    "./my-app",
			defaultOnEmpty: "./my-app",
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
			desc:     "Verify selections. [1-9] edits a line, enter generates.",
			kind:     kindReview,
		},
	}
}
