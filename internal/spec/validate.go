package spec

import (
	"errors"
	"fmt"

	"golang.org/x/mod/module"
)

// Validate checks that the spec is internally consistent: every enum field
// holds a legal value, the module path is a valid Go module path, and the
// dependency rules between optional features are satisfied. All violations
// are returned together so the caller can display them at once.
func (s ProjectSpec) Validate() error {
	var errs []error

	switch s.AppType {
	case AppWeb, AppAPI:
	default:
		errs = append(errs, fmt.Errorf("application type must be %q or %q", AppWeb, AppAPI))
	}

	switch s.Database {
	case DBNone, DBSQLite, DBPostgres, DBMySQL:
	default:
		errs = append(errs, fmt.Errorf("database must be one of %q, %q, %q or %q", DBNone, DBSQLite, DBPostgres, DBMySQL))
	}

	switch s.Router {
	case RouterServeMux, RouterFlow, RouterChi, RouterGorilla, RouterHttpRouter:
	default:
		errs = append(errs, fmt.Errorf("router must be one of %q, %q, %q, %q or %q",
			RouterServeMux, RouterFlow, RouterChi, RouterGorilla, RouterHttpRouter))
	}

	switch s.ConfigSource {
	case ConfigFlags, ConfigEnv:
	default:
		errs = append(errs, fmt.Errorf("config source must be %q or %q", ConfigFlags, ConfigEnv))
	}

	switch s.LogFormat {
	case LogPlaintext, LogJSON, LogColorized:
	default:
		errs = append(errs, fmt.Errorf("log format must be one of %q, %q or %q", LogPlaintext, LogJSON, LogColorized))
	}

	if err := module.CheckPath(s.ModulePath); err != nil {
		errs = append(errs, fmt.Errorf("module path %q is not a valid Go module path: %w", s.ModulePath, err))
	}

	// SQL migrations only make sense with a database selected.
	if s.Migrations && s.Database == DBNone {
		errs = append(errs, fmt.Errorf("sql migrations require a database"))
	}

	// Server-side sessions are a web-application feature and the session
	// data lives in the database.
	if s.Sessions {
		if s.AppType != AppWeb {
			errs = append(errs, fmt.Errorf("server-side sessions are only available for web applications"))
		}
		if s.Database == DBNone {
			errs = append(errs, fmt.Errorf("server-side sessions require a database"))
		}
	}

	// Custom error pages are HTML pages, so they only apply to web apps.
	if s.CustomErrorPages && s.AppType != AppWeb {
		errs = append(errs, fmt.Errorf("custom error pages are only available for web applications"))
	}

	// Error notifications are delivered by email.
	if s.ErrorNotifications && !s.Emails {
		errs = append(errs, fmt.Errorf("error notifications require the emails feature"))
	}

	// The full accounts variant is a web-application feature built on top of
	// sessions, the database and email sending.
	if s.AccountsFull {
		if s.AppType != AppWeb {
			errs = append(errs, fmt.Errorf("user accounts (full) are only available for web applications"))
		}
		if s.Database == DBNone {
			errs = append(errs, fmt.Errorf("user accounts (full) require a database"))
		}
		if !s.Sessions {
			errs = append(errs, fmt.Errorf("user accounts (full) require server-side sessions"))
		}
		if !s.Emails {
			errs = append(errs, fmt.Errorf("user accounts (full) require the emails feature"))
		}
	}

	// The JWT accounts variant is an API feature; activation tokens are
	// delivered by email.
	if s.AccountsJWT {
		if s.AppType != AppAPI {
			errs = append(errs, fmt.Errorf("user accounts (JWT) are only available for JSON APIs"))
		}
		if s.Database == DBNone {
			errs = append(errs, fmt.Errorf("user accounts (JWT) require a database"))
		}
		if !s.Emails {
			errs = append(errs, fmt.Errorf("user accounts (JWT) require the emails feature"))
		}
	}

	if s.AccountsFull && s.AccountsJWT {
		errs = append(errs, fmt.Errorf("only one user accounts variant can be selected"))
	}

	return joinErrors(errs)
}

// joinErrors returns nil for an empty slice, and a single error joining
// all violations (one per line) otherwise.
func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
