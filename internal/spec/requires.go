package spec

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
