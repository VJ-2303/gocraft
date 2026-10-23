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

// askCustomize shows the step 2 form: the database, router, configuration
// and optional-feature choices. Groups hidden by dependency rules (for
// example migrations without a database) are skipped automatically.
func askCustomize(s *spec.ProjectSpec) error {
	database := string(s.Database)
	router := string(s.Router)
	configSource := string(s.ConfigSource)
	logFormat := string(s.LogFormat)
	migrationsEnabled := s.Migrations
	sessions := s.Sessions

	var freeFeatures []string
	if s.RequestLogging {
		freeFeatures = append(freeFeatures, "request-logging")
	}
	if s.BasicAuth {
		freeFeatures = append(freeFeatures, "basic-auth")
	}
	if s.GitIntegration {
		freeFeatures = append(freeFeatures, "git")
	}
	if s.SecureCookies {
		freeFeatures = append(freeFeatures, "secure-cookies")
	}

	isWeb := s.AppType == spec.AppWeb
	hasDB := func() bool { return database != string(spec.DBNone) }

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Step 2 · Which database do you want to use?").
				Options(
					huh.NewOption("None", string(spec.DBNone)),
					huh.NewOption("SQLite", string(spec.DBSQLite)),
					huh.NewOption("PostgreSQL", string(spec.DBPostgres)),
					huh.NewOption("MySQL", string(spec.DBMySQL)),
				).
				Value(&database),
			huh.NewSelect[string]().
				Title("Which router do you want to use?").
				Options(
					huh.NewOption("http.ServeMux (standard library)", string(spec.RouterServeMux)),
					huh.NewOption("Flow", string(spec.RouterFlow)),
					huh.NewOption("Chi", string(spec.RouterChi)),
					huh.NewOption("Gorilla Mux", string(spec.RouterGorilla)),
					huh.NewOption("HttpRouter", string(spec.RouterHttpRouter)),
				).
				Value(&router),
			huh.NewSelect[string]().
				Title("Where should configuration settings be read from?").
				Options(
					huh.NewOption("Command-line flags", string(spec.ConfigFlags)),
					huh.NewOption("Environment variables", string(spec.ConfigEnv)),
				).
				Value(&configSource),
			huh.NewSelect[string]().
				Title("Which log format do you want to use?").
				Options(
					huh.NewOption("Plaintext", string(spec.LogPlaintext)),
					huh.NewOption("JSON", string(spec.LogJSON)),
					huh.NewOption("Colorized plaintext", string(spec.LogColorized)),
				).
				Value(&logFormat),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include helpers for managing SQL migrations?").
				Description("Adds an embedded migrations directory, Makefile targets, and automatic migration on startup.").
				Value(&migrationsEnabled),
		).WithHideFunc(func() bool { return !hasDB() }),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include server-side sessions?").
				Description("Session data is stored in your database via alexedwards/scs; the cookie holds only a random ID.").
				Value(&sessions),
		).WithHideFunc(func() bool { return !isWeb || !hasDB() }),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Any optional features?").
				Options(
					huh.NewOption("Request logging — log requests with status code and body size", "request-logging"),
					huh.NewOption("Basic auth — HTTP basic authentication middleware", "basic-auth"),
					huh.NewOption("Git integration — .gitignore and version from Git metadata", "git"),
					huh.NewOption("Secure cookies — signed and encrypted cookie helpers", "secure-cookies"),
				).
				Value(&freeFeatures),
		),
	).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return err
	}

	s.Database = spec.Database(database)
	s.Router = spec.Router(router)
	s.ConfigSource = spec.ConfigSource(configSource)
	s.LogFormat = spec.LogFormat(logFormat)
	s.Migrations = migrationsEnabled
	s.Sessions = isWeb && hasDB() && sessions
	s.RequestLogging = contains(freeFeatures, "request-logging")
	s.BasicAuth = contains(freeFeatures, "basic-auth")
	s.GitIntegration = contains(freeFeatures, "git")
	s.SecureCookies = contains(freeFeatures, "secure-cookies")

	return nil
}

// askPlusFeatures shows the step 3 form: the paid-tier feature choices.
// This tool generates all of them, free of charge.
func askPlusFeatures(s *spec.ProjectSpec) error {
	isWeb := s.AppType == spec.AppWeb

	var plusFeatures []string
	if s.AutoHTTPS {
		plusFeatures = append(plusFeatures, "auto-https")
	}
	if s.CustomErrorPages {
		plusFeatures = append(plusFeatures, "error-pages")
	}
	if s.Emails {
		plusFeatures = append(plusFeatures, "emails")
	}
	if s.Tests {
		plusFeatures = append(plusFeatures, "tests")
	}

	var emailFeatures []string
	if s.ErrorNotifications {
		emailFeatures = append(emailFeatures, "error-notifications")
	}

	accounts := "none"
	switch {
	case s.AccountsFull:
		accounts = "full"
	case s.AccountsJWT:
		accounts = "jwt"
	}

	emailsSelected := func() bool { return contains(plusFeatures, "emails") }

	accountsOptions := []huh.Option[string]{huh.NewOption("None", "none")}
	if isWeb {
		accountsOptions = append(accountsOptions,
			huh.NewOption("Full — signup, login, logout, password reset, CSRF protection", "full"))
	} else {
		accountsOptions = append(accountsOptions,
			huh.NewOption("JWT — email/password authentication with JSON Web Tokens", "jwt"))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Step 3 · Which Plus features do you want?").
				Description("Autostrada charges $29 for these; here they are included for free.").
				Options(func() []huh.Option[string] {
					options := []huh.Option[string]{
						huh.NewOption("Automatic HTTPS — Let's Encrypt TLS in production", "auto-https"),
					}
					if isWeb {
						options = append(options,
							huh.NewOption("Custom error pages — custom 400, 404 and 500 pages", "error-pages"))
					}
					options = append(options,
						huh.NewOption("Emails — email templates and SMTP sending helpers", "emails"),
						huh.NewOption("Tests — full test coverage and standardized test helpers", "tests"))
					return options
				}()...).
				Value(&plusFeatures),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Any email-dependent features?").
				Options(
					huh.NewOption("Error notifications — email an admin address on runtime errors", "error-notifications"),
				).
				Value(&emailFeatures),
		).WithHideFunc(func() bool { return !emailsSelected() }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("User accounts and authentication?").
				Options(accountsOptions...).
				Value(&accounts),
		),
	).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return err
	}

	s.AutoHTTPS = contains(plusFeatures, "auto-https")
	s.CustomErrorPages = isWeb && contains(plusFeatures, "error-pages")
	s.Emails = emailsSelected()
	s.ErrorNotifications = contains(emailFeatures, "error-notifications")
	s.Tests = contains(plusFeatures, "tests")

	switch accounts {
	case "full":
		s.AccountsFull, s.AccountsJWT = true, false
	case "jwt":
		s.AccountsFull, s.AccountsJWT = false, true
	default:
		s.AccountsFull, s.AccountsJWT = false, false
	}

	return nil
}

// review displays a summary of every choice and asks the user to confirm,
// returning true to proceed with generation.
func review(s *spec.ProjectSpec) (bool, error) {
	var proceed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Review — generate this codebase?").
				Description(reviewSummary(s)).
				Affirmative("Generate").
				Negative("Start over").
				Value(&proceed),
		),
	).WithShowHelp(true)

	if err := form.Run(); err != nil {
		return false, err
	}
	return proceed, nil
}

// reviewSummary renders the review screen body.
func reviewSummary(s *spec.ProjectSpec) string {
	row := func(label, value string) string {
		return fmt.Sprintf("%-18s %s\n", label, value)
	}
	rowBool := func(label string, v bool) string {
		return row(label, yesNo(v))
	}

	accounts := "none"
	switch {
	case s.AccountsFull:
		accounts = "full (web)"
	case s.AccountsJWT:
		accounts = "JWT (API)"
	}

	out := row("Application type", string(s.AppType))
	out += row("Module path", s.ModulePath)
	out += row("Database", string(s.Database))
	out += rowBool("SQL migrations", s.Migrations)
	out += row("Router", string(s.Router))
	out += row("Config source", string(s.ConfigSource))
	out += row("Log format", string(s.LogFormat))
	out += rowBool("Request logging", s.RequestLogging)
	out += rowBool("Basic auth", s.BasicAuth)
	out += rowBool("Git integration", s.GitIntegration)
	out += rowBool("Secure cookies", s.SecureCookies)
	out += rowBool("Sessions", s.Sessions)
	out += rowBool("Automatic HTTPS", s.AutoHTTPS)
	out += rowBool("Custom error pages", s.CustomErrorPages)
	out += rowBool("Emails", s.Emails)
	out += rowBool("Error notifications", s.ErrorNotifications)
	out += rowBool("Tests", s.Tests)
	out += row("User accounts", accounts)
	return out
}

// yesNo renders a bool as a friendly word.
func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
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

// contains reports whether list includes want.
func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
