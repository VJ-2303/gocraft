package ui

import (
	"fmt"
	"strings"

	"github.com/VJ-2303/gocraft/internal/spec"
)

// wizardValues holds all answer state during the interactive flow.
type wizardValues struct {
	appType          string
	modulePath       string
	database         string
	router           string
	configSource     string
	logFormat        string
	migrations       bool
	sessions         bool
	requestLogging   bool
	basicAuth        bool
	gitIntegration   bool
	secureCookies    bool
	autoHTTPS        bool
	customErrorPages bool
	emails           bool
	errorNotifs      bool
	tests            bool
	accounts         string
	targetDir        string
}

func defaultWizardValues() *wizardValues {
	return &wizardValues{
		appType:        string(spec.AppWeb),
		modulePath:     "example.com/myapp",
		database:       string(spec.DBNone),
		router:         string(spec.RouterServeMux),
		configSource:   string(spec.ConfigFlags),
		logFormat:      string(spec.LogPlaintext),
		gitIntegration: true,
		tests:          true,
		accounts:       "none",
		targetDir:      "./my-app",
	}
}

// toSpec converts wizard values to a ProjectSpec.
func (v *wizardValues) toSpec() spec.ProjectSpec {
	s := spec.ProjectSpec{
		AppType:            spec.AppType(v.appType),
		ModulePath:         v.modulePath,
		Database:           spec.Database(v.database),
		Migrations:         v.migrations,
		Router:             spec.Router(v.router),
		ConfigSource:       spec.ConfigSource(v.configSource),
		LogFormat:          spec.LogFormat(v.logFormat),
		RequestLogging:     v.requestLogging,
		BasicAuth:          v.basicAuth,
		GitIntegration:     v.gitIntegration,
		SecureCookies:      v.secureCookies,
		Sessions:           v.sessions,
		AutoHTTPS:          v.autoHTTPS,
		CustomErrorPages:   v.customErrorPages,
		Emails:             v.emails,
		ErrorNotifications: v.errorNotifs,
		Tests:              v.tests,
	}

	switch v.accounts {
	case "full":
		s.AccountsFull = true
	case "jwt":
		s.AccountsJWT = true
	}

	return s
}

// applyCascade auto-enables prerequisites when a feature is selected.
func (v *wizardValues) applyCascade(id string, val any) {
	switch id {
	case "accounts":
		if s, ok := val.(string); ok && s == "full" {
			v.sessions = true
			v.emails = true
			if v.database == string(spec.DBNone) {
				v.database = string(spec.DBSQLite)
			}
		} else if ok && s == "jwt" {
			v.emails = true
			if v.database == string(spec.DBNone) {
				v.database = string(spec.DBSQLite)
			}
		}
	case "sessions":
		if b, ok := val.(bool); ok && b {
			if v.database == string(spec.DBNone) {
				v.database = string(spec.DBSQLite)
			}
		}
	case "migrations":
		if b, ok := val.(bool); ok && b {
			if v.database == string(spec.DBNone) {
				v.database = string(spec.DBSQLite)
			}
		}
	case "error_notifications":
		if b, ok := val.(bool); ok && b {
			v.emails = true
		}
	}
}

// detectCascadeBadges returns live inline notification badges for options on the current question.
func (v *wizardValues) detectCascadeBadges(id string, optionValue string) string {
	switch id {
	case "accounts":
		if optionValue == "full" {
			var needs []string
			if v.database == string(spec.DBNone) {
				needs = append(needs, "SQLite")
			}
			if !v.sessions {
				needs = append(needs, "Sessions")
			}
			if !v.emails {
				needs = append(needs, "Emails")
			}
			if len(needs) > 0 {
				return fmt.Sprintf("+auto-enables: %s", formatList(needs))
			}
		} else if optionValue == "jwt" {
			var needs []string
			if v.database == string(spec.DBNone) {
				needs = append(needs, "SQLite")
			}
			if !v.emails {
				needs = append(needs, "Emails")
			}
			if len(needs) > 0 {
				return fmt.Sprintf("+auto-enables: %s", formatList(needs))
			}
		}
	case "sessions":
		if optionValue == "true" && v.database == string(spec.DBNone) {
			return "+auto-enables SQLite"
		}
	case "migrations":
		if optionValue == "true" && v.database == string(spec.DBNone) {
			return "+auto-enables SQLite"
		}
	case "error_notifications":
		if optionValue == "true" && !v.emails {
			return "+auto-enables Emails"
		}
	case "database":
		if optionValue == string(spec.DBNone) {
			var lost []string
			if v.migrations {
				lost = append(lost, "Migrations")
			}
			if v.sessions {
				lost = append(lost, "Sessions")
			}
			if v.accounts != "none" {
				lost = append(lost, "Accounts")
			}
			if len(lost) > 0 {
				return fmt.Sprintf("!will disable: %s", formatList(lost))
			}
		}
	case "app_type":
		if optionValue == string(spec.AppAPI) && v.appType == string(spec.AppWeb) {
			var lost []string
			if v.sessions {
				lost = append(lost, "Sessions")
			}
			if v.customErrorPages {
				lost = append(lost, "Error Pages")
			}
			if v.accounts == "full" {
				lost = append(lost, "Full Accounts")
			}
			if len(lost) > 0 {
				return fmt.Sprintf("!will disable web features: %s", formatList(lost))
			}
		}
	}
	return ""
}

// destructiveConflict checks if selecting an option destroys prior choices.
func (v *wizardValues) destructiveConflict(id string, optionValue string) (string, bool) {
	if id == "database" && optionValue == string(spec.DBNone) {
		var affected []string
		if v.migrations {
			affected = append(affected, "SQL Migrations")
		}
		if v.sessions {
			affected = append(affected, "Server-side Sessions")
		}
		if v.accounts != "none" {
			affected = append(affected, "User Accounts")
		}
		if len(affected) > 0 {
			return fmt.Sprintf("Selecting 'None' will disable active features: %s.", formatList(affected)), true
		}
	}

	if id == "app_type" && optionValue == string(spec.AppAPI) && v.appType == string(spec.AppWeb) {
		var affected []string
		if v.sessions {
			affected = append(affected, "Server-side Sessions")
		}
		if v.customErrorPages {
			affected = append(affected, "Custom Error Pages")
		}
		if v.accounts == "full" {
			affected = append(affected, "Full User Accounts")
		}
		if len(affected) > 0 {
			return fmt.Sprintf("Switching to JSON API will remove web-only features: %s.", formatList(affected)), true
		}
	}

	return "", false
}

func formatList(items []string) string {
	return strings.Join(items, ", ")
}
