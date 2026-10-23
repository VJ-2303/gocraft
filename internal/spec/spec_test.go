package spec

import (
	"errors"
	"strings"
	"testing"
)

// validWeb is a spec with every field holding a legal value.
func validWeb() ProjectSpec {
	return ProjectSpec{
		AppType:      AppWeb,
		ModulePath:   "github.com/alice/example",
		Database:     DBNone,
		Router:       RouterServeMux,
		ConfigSource: ConfigFlags,
		LogFormat:    LogPlaintext,
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*ProjectSpec)
		wantErr string // empty means expect success
	}{
		{
			name:   "valid defaults",
			mutate: func(s *ProjectSpec) {},
		},
		{
			name: "valid with every free option",
			mutate: func(s *ProjectSpec) {
				s.Database = DBPostgres
				s.Migrations = true
				s.Router = RouterChi
				s.ConfigSource = ConfigEnv
				s.LogFormat = LogJSON
				s.RequestLogging = true
				s.BasicAuth = true
				s.GitIntegration = true
				s.SecureCookies = true
				s.Sessions = true
			},
		},
		{
			name: "valid api with every applicable option",
			mutate: func(s *ProjectSpec) {
				s.AppType = AppAPI
				s.Database = DBMySQL
				s.Migrations = true
				s.Router = RouterHttpRouter
				s.ConfigSource = ConfigEnv
				s.LogFormat = LogColorized
				s.RequestLogging = true
				s.BasicAuth = true
				s.GitIntegration = true
				s.SecureCookies = true
				s.Emails = true
				s.ErrorNotifications = true
				s.Tests = true
				s.AccountsJWT = true
				s.AutoHTTPS = true
			},
		},
		{
			name: "valid web with every plus feature",
			mutate: func(s *ProjectSpec) {
				s.Database = DBSQLite
				s.Migrations = true
				s.Sessions = true
				s.Emails = true
				s.AutoHTTPS = true
				s.CustomErrorPages = true
				s.ErrorNotifications = true
				s.Tests = true
				s.AccountsFull = true
			},
		},
		{
			name:    "missing app type",
			mutate:  func(s *ProjectSpec) { s.AppType = "" },
			wantErr: "application type",
		},
		{
			name:    "invalid app type",
			mutate:  func(s *ProjectSpec) { s.AppType = "cli" },
			wantErr: "application type",
		},
		{
			name:    "invalid database",
			mutate:  func(s *ProjectSpec) { s.Database = "mongodb" },
			wantErr: "database",
		},
		{
			name:    "invalid router",
			mutate:  func(s *ProjectSpec) { s.Router = "gin" },
			wantErr: "router",
		},
		{
			name:    "invalid config source",
			mutate:  func(s *ProjectSpec) { s.ConfigSource = "yaml" },
			wantErr: "config source",
		},
		{
			name:    "invalid log format",
			mutate:  func(s *ProjectSpec) { s.LogFormat = "syslog" },
			wantErr: "log format",
		},
		{
			name:    "empty module path",
			mutate:  func(s *ProjectSpec) { s.ModulePath = "" },
			wantErr: "module path",
		},
		{
			name:    "module path with scheme",
			mutate:  func(s *ProjectSpec) { s.ModulePath = "https://github.com/alice/example" },
			wantErr: "module path",
		},
		{
			name:    "module path without host dot",
			mutate:  func(s *ProjectSpec) { s.ModulePath = "alice/example" },
			wantErr: "module path",
		},
		{
			name: "migrations without database",
			mutate: func(s *ProjectSpec) {
				s.Database = DBNone
				s.Migrations = true
			},
			wantErr: "migrations require a database",
		},
		{
			name: "sessions without database",
			mutate: func(s *ProjectSpec) {
				s.Sessions = true
			},
			wantErr: "sessions require a database",
		},
		{
			name: "sessions for api app",
			mutate: func(s *ProjectSpec) {
				s.AppType = AppAPI
				s.Database = DBSQLite
				s.Sessions = true
			},
			wantErr: "only available for web applications",
		},
		{
			name: "custom error pages for api app",
			mutate: func(s *ProjectSpec) {
				s.AppType = AppAPI
				s.CustomErrorPages = true
			},
			wantErr: "only available for web applications",
		},
		{
			name: "error notifications without emails",
			mutate: func(s *ProjectSpec) {
				s.ErrorNotifications = true
			},
			wantErr: "error notifications require the emails feature",
		},
		{
			name: "accounts full without sessions",
			mutate: func(s *ProjectSpec) {
				s.Database = DBSQLite
				s.Emails = true
				s.AccountsFull = true
			},
			wantErr: "user accounts (full) require server-side sessions",
		},
		{
			name: "accounts full without database",
			mutate: func(s *ProjectSpec) {
				s.Sessions = true
				s.Database = DBNone
				s.Emails = true
				s.AccountsFull = true
			},
			wantErr: "require a database",
		},
		{
			name: "accounts full for api app",
			mutate: func(s *ProjectSpec) {
				s.AppType = AppAPI
				s.Database = DBSQLite
				s.Emails = true
				s.AccountsFull = true
			},
			wantErr: "user accounts (full) are only available for web applications",
		},
		{
			name: "accounts jwt for web app",
			mutate: func(s *ProjectSpec) {
				s.Database = DBSQLite
				s.Emails = true
				s.AccountsJWT = true
			},
			wantErr: "user accounts (JWT) are only available for JSON APIs",
		},
		{
			name: "accounts jwt without emails",
			mutate: func(s *ProjectSpec) {
				s.AppType = AppAPI
				s.Database = DBSQLite
				s.AccountsJWT = true
			},
			wantErr: "user accounts (JWT) require the emails feature",
		},
		{
			name: "both account variants",
			mutate: func(s *ProjectSpec) {
				s.Database = DBSQLite
				s.Emails = true
				s.Sessions = true
				s.AccountsFull = true
				s.AppType = AppAPI
				s.AccountsJWT = true
			},
			wantErr: "only one user accounts variant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := validWeb()
			tt.mutate(&s)

			err := s.Validate()

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected valid spec, got error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("want error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestValidateReportsAllViolations(t *testing.T) {
	s := validWeb()
	s.Migrations = true // requires database
	s.ErrorNotifications = true

	err := s.Validate()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !strings.Contains(err.Error(), "migrations require a database") {
		t.Errorf("want migrations violation in %q", err.Error())
	}
	if !strings.Contains(err.Error(), "error notifications require the emails feature") {
		t.Errorf("want error notifications violation in %q", err.Error())
	}
}

func TestApplyDependencies(t *testing.T) {
	tests := []struct {
		name  string
		give  ProjectSpec
		check func(t *testing.T, got ProjectSpec)
	}{
		{
			name: "accounts full enables sessions and emails",
			give: ProjectSpec{AppType: AppWeb, ModulePath: "example.com/a", Database: DBSQLite, AccountsFull: true},
			check: func(t *testing.T, got ProjectSpec) {
				if !got.Sessions || !got.Emails {
					t.Fatalf("want sessions and emails enabled, got %+v", got)
				}
			},
		},
		{
			name: "accounts full defaults database to sqlite",
			give: ProjectSpec{AppType: AppWeb, ModulePath: "example.com/a", AccountsFull: true},
			check: func(t *testing.T, got ProjectSpec) {
				if got.Database != DBSQLite {
					t.Fatalf("want sqlite database, got %q", got.Database)
				}
			},
		},
		{
			name: "accounts jwt enables emails",
			give: ProjectSpec{AppType: AppAPI, ModulePath: "example.com/a", Database: DBMySQL, AccountsJWT: true},
			check: func(t *testing.T, got ProjectSpec) {
				if !got.Emails {
					t.Fatal("want emails enabled")
				}
			},
		},
		{
			name: "error notifications enables emails",
			give: ProjectSpec{AppType: AppAPI, ModulePath: "example.com/a", ErrorNotifications: true},
			check: func(t *testing.T, got ProjectSpec) {
				if !got.Emails {
					t.Fatal("want emails enabled")
				}
			},
		},
		{
			name: "sessions without database defaults to sqlite",
			give: ProjectSpec{AppType: AppWeb, ModulePath: "example.com/a", Sessions: true},
			check: func(t *testing.T, got ProjectSpec) {
				if got.Database != DBSQLite {
					t.Fatalf("want sqlite database, got %q", got.Database)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.give
			got.ApplyDependencies()
			tt.check(t, got)
		})
	}
}

func TestResolve(t *testing.T) {
	t.Run("rejects invalid spec", func(t *testing.T) {
		s := validWeb()
		s.ModulePath = "not a module path"

		if _, err := Resolve(s); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("defaults", func(t *testing.T) {
		r, err := Resolve(validWeb())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if r.CmdDir != "web" {
			t.Errorf("want CmdDir web, got %q", r.CmdDir)
		}
		if r.ProjectName != "example" {
			t.Errorf("want ProjectName example, got %q", r.ProjectName)
		}
		if r.UseDB || r.NeedsDB {
			t.Error("want no database for a plain web app")
		}
		// A plain web app requires only the form decoder; APIs with no
		// options require nothing.
		if len(r.Requires) != 1 || r.Requires[0].Path != "github.com/go-playground/form/v4" {
			t.Errorf("want only the form decoder requirement for a plain web app, got %v", r.Requires)
		}
		if r.Port < 1000 || r.Port > 9999 {
			t.Errorf("want port in [1000,9999], got %d", r.Port)
		}
		if !strings.HasPrefix(r.SessionCookieName, "session_") {
			t.Errorf("want session_ prefix on cookie name, got %q", r.SessionCookieName)
		}
	})

	t.Run("module path without slash uses host as project name", func(t *testing.T) {
		s := validWeb()
		s.ModulePath = "example.com"

		r, err := Resolve(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.ProjectName != "example.com" {
			t.Errorf("want example.com, got %q", r.ProjectName)
		}
	})

	t.Run("api app derives cmd dir", func(t *testing.T) {
		s := validWeb()
		s.AppType = AppAPI

		r, err := Resolve(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.CmdDir != "api" {
			t.Errorf("want CmdDir api, got %q", r.CmdDir)
		}
	})

	t.Run("postgres driver details and requires", func(t *testing.T) {
		s := validWeb()
		s.Database = DBPostgres
		s.Migrations = true
		s.Router = RouterFlow
		s.Sessions = true

		r, err := Resolve(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if r.Driver.Name != "pgx" || r.Driver.ImportPath != "github.com/jackc/pgx/v5/stdlib" {
			t.Errorf("unexpected driver info: %+v", r.Driver)
		}
		if !r.NeedsSessionStore {
			t.Error("want session store")
		}

		var paths []string
		for _, req := range r.Requires {
			paths = append(paths, req.Path)
		}
		for _, want := range []string{
			"github.com/alexedwards/flow",
			"github.com/jmoiron/sqlx",
			"github.com/jackc/pgx/v5",
			"github.com/alexedwards/scs/v2",
		} {
			if !contains(paths, want) {
				t.Errorf("want requirement %s, got %v", want, paths)
			}
		}
	})

	t.Run("duplicate requirements are removed", func(t *testing.T) {
		s := validWeb()
		s.AppType = AppAPI
		s.Database = DBSQLite
		s.Emails = true
		s.AccountsJWT = true
		s.AutoHTTPS = true

		r, err := Resolve(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		seen := make(map[string]int)
		for _, req := range r.Requires {
			seen[req.Path]++
		}
		for path, count := range seen {
			if count > 1 {
				t.Errorf("requirement %s appears %d times", path, count)
			}
		}
	})

	t.Run("needs mailer without emails option", func(t *testing.T) {
		s := validWeb()
		s.AppType = AppAPI
		s.Database = DBSQLite
		s.Emails = true

		r, err := Resolve(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !r.NeedsMailer {
			t.Error("want NeedsMailer")
		}
	})
}

func TestResolveErrorsAreInspectable(t *testing.T) {
	s := validWeb()
	s.Sessions = true // no database

	_, err := Resolve(s)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatalf("want a joinable error, got %T", err)
	}
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
