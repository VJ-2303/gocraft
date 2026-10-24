package spec

import (
	"fmt"
	"math/rand/v2"
	"path"
	"strings"
)

// GoVersion is the Go toolchain version that generated codebases target.
const GoVersion = "1.27"

// DriverInfo describes the database driver for a dialect.
type DriverInfo struct {
	// Name is the driver name registered with database/sql, e.g. "pgx".
	Name string

	// ModPath is the module path to list in go.mod.
	ModPath string

	// ImportPath is the package whose side effects register the driver.
	ImportPath string

	// ExampleDSN is a sample DSN for the dialect, shown in comments and the
	// generated README.
	ExampleDSN string
}

// drivers maps each SQL dialect to its driver details.
var drivers = map[Database]DriverInfo{
	DBSQLite: {
		Name:       "sqlite",
		ModPath:    "modernc.org/sqlite",
		ImportPath: "modernc.org/sqlite",
		ExampleDSN: "db.sqlite?_pragma=foreign_keys(1)",
	},
	DBPostgres: {
		Name:       "pgx",
		ModPath:    "github.com/jackc/pgx/v5",
		ImportPath: "github.com/jackc/pgx/v5/stdlib",
		ExampleDSN: "user:pass@localhost:5432/example?sslmode=disable&default_query_exec_mode=simple_protocol",
	},
	DBMySQL: {
		Name:       "mysql",
		ModPath:    "github.com/go-sql-driver/mysql",
		ImportPath: "github.com/go-sql-driver/mysql",
		ExampleDSN: "user:pass@tcp(localhost:3306)/example?parseTime=true&multiStatements=true",
	},
}

// requireVersions pins the module versions written into generated go.mod
// files. The generated Makefile's `tidy` and `upgradeable` targets manage
// these going forward.
var requireVersions = map[string]string{
	"github.com/jmoiron/sqlx":             "v1.4.0",
	"modernc.org/sqlite":                  "v1.61.0",
	"github.com/jackc/pgx/v5":             "v5.11.0",
	"github.com/go-sql-driver/mysql":      "v1.10.1",
	"github.com/alexedwards/flow":         "v1.2.0",
	"github.com/go-chi/chi/v5":            "v5.5.0",
	"github.com/gorilla/mux":              "v1.8.1",
	"github.com/julienschmidt/httprouter": "v1.3.0",
	"github.com/alexedwards/scs/v2":       "v2.9.0",
	"github.com/justinas/nosurf":          "v1.2.0",
	"github.com/golang-jwt/jwt/v5":        "v5.3.1",
	"github.com/go-playground/form/v4":    "v4.5.0",
	"github.com/wneessen/go-mail":         "v0.8.1",
	"golang.org/x/crypto":                 "v0.58.0",
}

// ResolvedSpec is the template view-model: the raw wizard answers plus
// every derived value the templates need. It is computed once and passed to
// every file render.
type ResolvedSpec struct {
	ProjectSpec

	// ProjectName is the human-friendly name of the project, taken from the
	// final element of the module path.
	ProjectName string

	// CmdDir is the directory under cmd/ holding the entrypoint: "web" or
	// "api".
	CmdDir string

	// UseDB reports whether a database was explicitly selected.
	UseDB bool

	// NeedsDB reports whether database integration is required, either
	// explicitly or because a feature like sessions or user accounts needs
	// somewhere to store data.
	NeedsDB bool

	// Dialect is the effective SQL dialect; DBNone when NeedsDB is false.
	Dialect Database

	// Driver holds driver details for the dialect; the zero value when
	// NeedsDB is false.
	Driver DriverInfo

	// MigrateTags are the golang-migrate build tags matching the dialect,
	// used by the Makefile migration targets.
	MigrateTags string

	// GoVersion is the Go toolchain version written into the generated
	// go.mod.
	GoVersion string

	// NeedsMailer reports whether SMTP sending is required, either because
	// the emails feature is selected or because a feature like error
	// notifications or user accounts sends mail.
	NeedsMailer bool

	// NeedsSessionStore reports whether the session manager is backed by a
	// database store (always true when sessions are enabled, since sessions
	// require a database).
	NeedsSessionStore bool

	// HasAuth reports whether either user accounts variant is enabled.
	HasAuth bool

	// Port is the default HTTP port for the codebase, randomized per
	// codebase so local apps don't collide.
	Port int

	// SessionCookieName is the default session cookie name, randomized per
	// codebase to avoid cookie clashes between local apps.
	SessionCookieName string

	// BasicAuthHashedPassword is the bcrypt hash of the default basic
	// auth password ("pa55word"), shipped as the flag default like
	// Autostrada's generated codebases.
	BasicAuthHashedPassword string

	// CookieSecretKey is a random 32-character string used by the secure
	// cookies helpers, generated per codebase for development. Documented
	// for replacement in production.
	CookieSecretKey string

	// Requires lists the module requirements (path + version) implied by
	// the selected options, for go.mod.
	Requires []RequireLine
}

// RequireLine is one entry in a generated go.mod require block.
type RequireLine struct {
	Path    string
	Version string
}

// ApplyDependencies automatically enables any optional feature that other
// selected features depend on. The wizard calls this whenever an answer
// changes, so the resulting spec always satisfies the dependency rules.
func (s *ProjectSpec) ApplyDependencies() {
	// Normalize the zero-value database to an explicit choice.
	if s.Database == "" {
		s.Database = DBNone
	}

	if s.AccountsFull {
		s.Sessions = true
		s.Emails = true
	}
	if s.AccountsJWT {
		s.Emails = true
	}
	if s.ErrorNotifications {
		s.Emails = true
	}
	if (s.Sessions || s.AccountsJWT) && s.Database == DBNone {
		s.Database = DBSQLite
	}
}

// Resolve validates the spec and computes the derived view-model fields.
// The random port and session cookie name are generated fresh on each call;
// callers that need deterministic output (tests, reproducible runs) can set
// them on the result afterwards.
func Resolve(s ProjectSpec) (ResolvedSpec, error) {
	if err := s.Validate(); err != nil {
		return ResolvedSpec{}, err
	}

	r := ResolvedSpec{
		ProjectSpec:             s,
		UseDB:                   s.Database != DBNone,
		NeedsSessionStore:       s.Sessions,
		HasAuth:                 s.AccountsFull || s.AccountsJWT,
		NeedsMailer:             s.Emails || s.ErrorNotifications || s.AccountsFull || s.AccountsJWT,
		Port:                    1000 + rand.IntN(9000),
		SessionCookieName:       "session_" + randomSuffix(8),
		BasicAuthHashedPassword: BasicAuthHashedPassword,
		CookieSecretKey:         randomKey(),
	}

	r.NeedsDB = r.UseDB || s.Sessions || s.AccountsFull || s.AccountsJWT
	if !r.NeedsDB {
		r.Dialect = DBNone
	} else {
		r.Dialect = s.Database
		if r.Dialect == DBNone {
			// A feature needs a database but none was chosen; default to
			// SQLite, which needs no server process.
			r.Dialect = DBSQLite
		}
		r.Driver = drivers[r.Dialect]
	}

	switch s.AppType {
	case AppAPI:
		r.CmdDir = "api"
	default:
		r.CmdDir = "web"
	}

	r.ProjectName = projectName(s.ModulePath)
	r.MigrateTags = migrateTags[r.Dialect]
	r.GoVersion = GoVersion
	r.Requires = s.requires(&r)

	return r, nil
}

// migrateTags maps each SQL dialect to its golang-migrate build tag.
var migrateTags = map[Database]string{
	DBSQLite:   "sqlite",
	DBPostgres: "postgres",
	DBMySQL:    "mysql",
}

// requires computes the go.mod requirement lines implied by the options.
func (s ProjectSpec) requires(r *ResolvedSpec) []RequireLine {
	var lines []RequireLine
	add := func(modPath string) {
		lines = append(lines, RequireLine{Path: modPath, Version: requireVersions[modPath]})
	}

	switch s.Router {
	case RouterFlow:
		add("github.com/alexedwards/flow")
	case RouterChi:
		add("github.com/go-chi/chi/v5")
	case RouterGorilla:
		add("github.com/gorilla/mux")
	case RouterHttpRouter:
		add("github.com/julienschmidt/httprouter")
	}

	if r.NeedsDB {
		add("github.com/jmoiron/sqlx")
		add(r.Driver.ModPath)
	}

	if s.Sessions {
		add("github.com/alexedwards/scs/v2")
	}

	if s.AppType == AppWeb {
		add("github.com/go-playground/form/v4")
	}

	if s.AccountsFull {
		add("github.com/justinas/nosurf")
	}

	if s.AccountsFull || s.AccountsJWT || s.BasicAuth {
		add("golang.org/x/crypto")
	}

	if r.NeedsMailer {
		add("github.com/wneessen/go-mail")
	}

	if s.AutoHTTPS {
		add("golang.org/x/crypto")
	}

	// Deduplicate (golang.org/x/crypto can be required by more than one
	// feature) and keep the order stable for reproducible output.
	seen := make(map[string]bool, len(lines))
	deduped := lines[:0]
	for _, line := range lines {
		if !seen[line.Path] {
			seen[line.Path] = true
			deduped = append(deduped, line)
		}
	}

	return deduped
}

// projectName extracts a human-friendly project name from a module path,
// e.g. "example" from "github.com/alice/example".
func projectName(modulePath string) string {
	if i := strings.LastIndex(modulePath, "/"); i >= 0 && i < len(modulePath)-1 {
		return path.Base(modulePath)
	}
	return modulePath
}

// randomSuffix returns n random lowercase alphanumeric characters.
func randomSuffix(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rand.IntN(len(alphabet))]
	}
	return string(b)
}

// randomKey returns a random 32-character alphanumeric string, suitable
// as a cookie signing or encryption secret for development.
func randomKey() string {
	return randomSuffix(32)
}

// BasicAuthHashedPassword is the bcrypt hash (cost 12) of the default
// basic auth password "pa55word".
const BasicAuthHashedPassword = "$2a$12$Oz6h/J3bLOmVNgl.6ivcFORe.oAIGvYOYXavY9VV50T3lMwt4dC.a"

// String returns a compact one-line summary of the spec, useful for logs.
func (r ResolvedSpec) String() string {
	return fmt.Sprintf("%s app %q module=%s db=%s router=%s config=%s log=%s",
		r.AppType, r.ProjectName, r.ModulePath, r.Dialect, r.Router, r.ConfigSource, r.LogFormat)
}
