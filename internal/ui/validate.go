package ui

import (
	"fmt"
	"os"

	"golang.org/x/mod/module"

	"github.com/vj/nomoreboiler/internal/spec"
)

// checkModulePath validates a Go module path.
func checkModulePath(value string) error {
	return module.CheckPath(value)
}

// checkTargetDir validates the target directory: it must not exist, or
// must exist and be empty.
func checkTargetDir(value string) error {
	entries, err := os.ReadDir(value)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("cannot use %s: %v", value, err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("directory exists and is not empty")
	}
	return nil
}

// apply converts the wizard answers into a resolved spec, auto-enabling
// any features the selected features depend on.
func apply(v *wizardValues) (spec.ResolvedSpec, error) {
	s := spec.ProjectSpec{
		AppType:            spec.AppType(v.appType),
		ModulePath:         v.modulePath,
		Database:           spec.Database(v.database),
		Router:             spec.Router(v.router),
		ConfigSource:       spec.ConfigSource(v.configSource),
		LogFormat:          spec.LogFormat(v.logFormat),
		Migrations:         v.migrations,
		Sessions:           v.sessions,
		RequestLogging:     v.requestLogging,
		BasicAuth:          v.basicAuth,
		GitIntegration:     v.gitIntegration,
		SecureCookies:      v.secureCookies,
		AutoHTTPS:          v.autoHTTPS,
		CustomErrorPages:   v.customErrorPages,
		Emails:             v.emails,
		ErrorNotifications: v.errorNotifications,
		Tests:              v.tests,
		AccountsFull:       v.accounts == "full",
		AccountsJWT:        v.accounts == "jwt",
	}

	s.ApplyDependencies()

	if err := s.Validate(); err != nil {
		return spec.ResolvedSpec{}, err
	}

	return spec.Resolve(s)
}

// yesNo renders a bool as a friendly word.
func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
