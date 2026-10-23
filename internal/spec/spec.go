// Package spec defines the set of options that describe a codebase to
// generate, validates them against the dependency rules between options, and
// resolves them into the view-model used by the template engine.
package spec

// AppType identifies the kind of application to generate.
type AppType string

const (
	// AppWeb is a traditional web application with server-side HTML
	// rendering.
	AppWeb AppType = "web"

	// AppAPI is a JSON API that processes JSON requests and returns JSON
	// responses.
	AppAPI AppType = "api"
)

// Database identifies the SQL database to integrate.
type Database string

const (
	DBNone     Database = "none"
	DBSQLite   Database = "sqlite"
	DBPostgres Database = "postgres"
	DBMySQL    Database = "mysql"
)

// Router identifies the HTTP router to wire the routes through.
type Router string

const (
	RouterServeMux   Router = "servemux"
	RouterFlow       Router = "flow"
	RouterChi        Router = "chi"
	RouterGorilla    Router = "gorilla"
	RouterHttpRouter Router = "httprouter"
)

// ConfigSource identifies where configuration settings are read from.
type ConfigSource string

const (
	ConfigFlags ConfigSource = "flags"
	ConfigEnv   ConfigSource = "env"
)

// LogFormat identifies the output format of the application logger.
type LogFormat string

const (
	LogPlaintext LogFormat = "plaintext"
	LogJSON      LogFormat = "json"
	LogColorized LogFormat = "colorized"
)

// ProjectSpec holds the user's answers from the wizard: every option that
// controls which files are generated and what they contain.
type ProjectSpec struct {
	// AppType is the kind of application to generate. Required.
	AppType AppType

	// ModulePath is the Go module path of the generated codebase, e.g.
	// "github.com/alice/example". Required.
	ModulePath string

	// Database is the SQL database to integrate, or DBNone.
	Database Database

	// Migrations includes helpers for managing SQL migrations: an embedded
	// migration directory, Makefile targets and startup auto-migration.
	Migrations bool

	// Router is the HTTP router to use.
	Router Router

	// ConfigSource controls whether settings are read from command-line
	// flags or environment variables.
	ConfigSource ConfigSource

	// LogFormat is the slog output format.
	LogFormat LogFormat

	// RequestLogging includes middleware that logs each request with the
	// response status code, body size and duration.
	RequestLogging bool

	// BasicAuth includes HTTP basic authentication middleware.
	BasicAuth bool

	// GitIntegration adds a .gitignore file and derives the application
	// version from Git metadata.
	GitIntegration bool

	// SecureCookies includes helpers for setting signed (and optionally
	// encrypted) cookies.
	SecureCookies bool

	// Sessions includes server-side session support, backed by the
	// database. Web applications only; requires a database.
	Sessions bool

	// AutoHTTPS provisions TLS certificates from Let's Encrypt and
	// redirects HTTP traffic to HTTPS in production.
	AutoHTTPS bool

	// CustomErrorPages renders custom 400, 404 and 500 error pages.
	// Web applications only.
	CustomErrorPages bool

	// Emails includes email templates and SMTP sending helpers.
	Emails bool

	// ErrorNotifications sends an email to an admin address whenever a
	// runtime error occurs. Requires emails.
	ErrorNotifications bool

	// Tests generates a full test suite for the application, including
	// standardized test helpers.
	Tests bool

	// AccountsFull includes user accounts with signup, login, logout,
	// password reset, authentication middleware and CSRF protection.
	// Web applications only; requires a database, sessions and emails.
	AccountsFull bool

	// AccountsJWT includes email/password authentication using JSON Web
	// Tokens. JSON APIs only; requires a database and emails.
	AccountsJWT bool
}

// Defaults returns a ProjectSpec populated with the wizard's default
// selections.
func Defaults() ProjectSpec {
	return ProjectSpec{
		AppType:      AppWeb,
		Database:     DBNone,
		Migrations:   false,
		Router:       RouterServeMux,
		ConfigSource: ConfigFlags,
		LogFormat:    LogPlaintext,
	}
}
