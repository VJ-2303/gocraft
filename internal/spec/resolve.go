package spec

import (
	"fmt"
	"math/rand/v2"
	"path"
	"strings"
)

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
	// auth password ("pa55word"), shipped as the flag default for
	// out-of-the-box local testing.
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

// String returns a compact one-line summary of the spec, useful for logs.
func (r ResolvedSpec) String() string {
	return fmt.Sprintf("%s app %q module=%s db=%s router=%s config=%s log=%s",
		r.AppType, r.ProjectName, r.ModulePath, r.Dialect, r.Router, r.ConfigSource, r.LogFormat)
}
