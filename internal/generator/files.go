package generator

import (
	"fmt"

	"github.com/vj/nomoreboiler/internal/spec"
)

// FileSpec selects one file for generation. Template is the path of the
// embedded template or asset; OutPath is the destination path relative to
// the target directory (itself a template, so it can depend on the spec).
//
// Copy files are written verbatim from the embedded FS; they are typically
// HTML or email templates that contain their own Go template syntax and
// must not be rendered a second time. Rendered files are executed with the
// resolved spec as their data.
type FileSpec struct {
	Template string
	OutPath  string
	Cond     bool
	Copy     bool
}

// fileMatrix computes the ordered list of files to generate for a resolved
// spec. Files are listed in a stable order: root files first, then
// assets/, then internal/, then cmd/.
func fileMatrix(rsp spec.ResolvedSpec) []FileSpec {
	files := []FileSpec{
		{Template: "common/gomod.tmpl", OutPath: "go.mod", Cond: true},
		{Template: "common/gitignore.tmpl", OutPath: ".gitignore", Cond: rsp.GitIntegration},
		{Template: "common/makefile.tmpl", OutPath: "Makefile", Cond: true},
		{Template: "common/readme.tmpl", OutPath: "README.md", Cond: true},

		// Shared internal packages used by both application types.
		{Template: "shared/internal/version/version.go.tmpl", OutPath: "internal/version/version.go", Cond: true},
		{Template: "shared/internal/funcs/funcs.go.tmpl", OutPath: "internal/funcs/funcs.go", Cond: true},
		{Template: "shared/internal/response/json.go.tmpl", OutPath: "internal/response/json.go", Cond: true},
		{Template: "shared/internal/database/database.go.tmpl", OutPath: "internal/database/database.go", Cond: rsp.NeedsDB},
		{Template: "shared/internal/database/users.go.tmpl", OutPath: "internal/database/users.go", Cond: rsp.HasAuth},
		{Template: "shared/internal/database/password_resets.go.tmpl", OutPath: "internal/database/password_resets.go", Cond: rsp.AccountsFull},
		{Template: "shared/internal/database/activation_tokens.go.tmpl", OutPath: "internal/database/activation_tokens.go", Cond: rsp.AccountsJWT},
		{Template: "shared/internal/auth/jwt.go.tmpl", OutPath: "internal/auth/jwt.go", Cond: rsp.AccountsJWT},
		{Template: "shared/internal/request/json.go.tmpl", OutPath: "internal/request/json.go", Cond: rsp.AppType == spec.AppAPI},
		{Template: "shared/internal/env/env.go.tmpl", OutPath: "internal/env/env.go", Cond: rsp.ConfigSource == spec.ConfigEnv},
		{Template: "shared/internal/log/colorized.go.tmpl", OutPath: "internal/log/colorized.go", Cond: rsp.LogFormat == spec.LogColorized},

		// The assets package embeds migrations (both app types) and the
		// static files and HTML templates (web apps only).
		{Template: "shared/assets/efs.go.tmpl", OutPath: "assets/efs.go", Cond: rsp.AppType == spec.AppWeb || rsp.Migrations},
		{Template: "shared/assets/migrations/000001_init.up.sql", OutPath: "assets/migrations/000001_init.up.sql", Cond: rsp.Migrations, Copy: true},
		{Template: "shared/assets/migrations/000001_init.down.sql", OutPath: "assets/migrations/000001_init.down.sql", Cond: rsp.Migrations, Copy: true},
		{Template: "shared/assets/migrations/000002_create_sessions_table.up.sql.tmpl", OutPath: "assets/migrations/000002_create_sessions_table.up.sql", Cond: rsp.Sessions},
		{Template: "shared/assets/migrations/000002_create_sessions_table.down.sql.tmpl", OutPath: "assets/migrations/000002_create_sessions_table.down.sql", Cond: rsp.Sessions},
		{Template: "shared/assets/migrations/000003_create_users_table.up.sql.tmpl", OutPath: "assets/migrations/000003_create_users_table.up.sql", Cond: rsp.HasAuth},
		{Template: "shared/assets/migrations/000003_create_users_table.down.sql.tmpl", OutPath: "assets/migrations/000003_create_users_table.down.sql", Cond: rsp.HasAuth},
		{Template: "shared/assets/migrations/000004_create_password_resets_table.up.sql.tmpl", OutPath: "assets/migrations/000004_create_password_resets_table.up.sql", Cond: rsp.AccountsFull},
		{Template: "shared/assets/migrations/000004_create_password_resets_table.down.sql.tmpl", OutPath: "assets/migrations/000004_create_password_resets_table.down.sql", Cond: rsp.AccountsFull},
		{Template: "shared/assets/migrations/000004_create_activation_tokens_table.up.sql.tmpl", OutPath: "assets/migrations/000004_create_activation_tokens_table.up.sql", Cond: rsp.AccountsJWT},
		{Template: "shared/assets/migrations/000004_create_activation_tokens_table.down.sql.tmpl", OutPath: "assets/migrations/000004_create_activation_tokens_table.down.sql", Cond: rsp.AccountsJWT},

		{Template: "shared/internal/cookies/cookies.go.tmpl", OutPath: "internal/cookies/cookies.go", Cond: rsp.SecureCookies},
		{Template: "shared/internal/mailer/mailer.go.tmpl", OutPath: "internal/mailer/mailer.go", Cond: rsp.NeedsMailer},
		{Template: "shared/internal/password/hash.go.tmpl", OutPath: "internal/password/hash.go", Cond: rsp.HasAuth},
		{Template: "shared/internal/password/common.go.tmpl", OutPath: "internal/password/common.go", Cond: rsp.HasAuth},
		{Template: "shared/internal/validator/validator.go.tmpl", OutPath: "internal/validator/validator.go", Cond: rsp.HasAuth},
		{Template: "shared/internal/request/postform.go.tmpl", OutPath: "internal/request/postform.go", Cond: rsp.AccountsFull},

		{Template: "shared/assets/emails/example.tmpl", OutPath: "assets/emails/example.tmpl", Cond: rsp.NeedsMailer, Copy: true},
		{Template: "shared/assets/emails/error-notification.tmpl", OutPath: "assets/emails/error-notification.tmpl", Cond: rsp.ErrorNotifications, Copy: true},
		{Template: "web/assets/emails/forgot-password.tmpl", OutPath: "assets/emails/forgot-password.tmpl", Cond: rsp.AccountsFull, Copy: true},
		{Template: "shared/assets/emails/account-activation.tmpl", OutPath: "assets/emails/account-activation.tmpl", Cond: rsp.AccountsJWT, Copy: true},
	}

	files = append(files, appFiles(rsp)...)

	selected := files[:0]
	for _, f := range files {
		if f.Cond {
			selected = append(selected, f)
		}
	}
	return selected
}

// appFiles returns the application-type-specific files (web or api).
func appFiles(rsp spec.ResolvedSpec) []FileSpec {
	if rsp.AppType == spec.AppAPI {
		return apiFiles(rsp)
	}
	return webFiles(rsp)
}

// webFiles returns the files for a traditional web application.
func webFiles(rsp spec.ResolvedSpec) []FileSpec {
	routes := fmt.Sprintf("web/cmd/web/routes/%s.go.tmpl", rsp.Router)
	return []FileSpec{
		{Template: "shared/internal/response/html.go.tmpl", OutPath: "internal/response/html.go", Cond: true},

		{Template: routes, OutPath: "cmd/web/routes.go", Cond: true},

		{Template: "web/assets/static/css/main.css", OutPath: "assets/static/css/main.css", Cond: true, Copy: true},
		{Template: "web/assets/html/base.tmpl", OutPath: "assets/html/base.tmpl", Cond: true, Copy: true},
		{Template: "web/assets/html/components/nav-auth.tmpl", OutPath: "assets/html/components/nav.tmpl", Cond: rsp.AccountsFull, Copy: true},
		{Template: "web/assets/html/components/nav.tmpl", OutPath: "assets/html/components/nav.tmpl", Cond: !rsp.AccountsFull, Copy: true},
		{Template: "web/assets/html/pages/home.tmpl", OutPath: "assets/html/pages/home.tmpl", Cond: true, Copy: true},
		{Template: "web/assets/html/pages/error.tmpl", OutPath: "assets/html/pages/error.tmpl", Cond: rsp.CustomErrorPages, Copy: true},
		{Template: "web/assets/html/pages/signup.tmpl", OutPath: "assets/html/pages/signup.tmpl", Cond: rsp.AccountsFull, Copy: true},
		{Template: "web/assets/html/pages/login.tmpl", OutPath: "assets/html/pages/login.tmpl", Cond: rsp.AccountsFull, Copy: true},
		{Template: "web/assets/html/pages/restricted.tmpl", OutPath: "assets/html/pages/restricted.tmpl", Cond: rsp.AccountsFull, Copy: true},
		{Template: "web/assets/html/pages/forgotten-password.tmpl", OutPath: "assets/html/pages/forgotten-password.tmpl", Cond: rsp.AccountsFull, Copy: true},
		{Template: "web/assets/html/pages/forgotten-password-confirmation.tmpl", OutPath: "assets/html/pages/forgotten-password-confirmation.tmpl", Cond: rsp.AccountsFull, Copy: true},
		{Template: "web/assets/html/pages/password-reset.tmpl", OutPath: "assets/html/pages/password-reset.tmpl", Cond: rsp.AccountsFull, Copy: true},
		{Template: "web/assets/html/pages/password-reset-confirmation.tmpl", OutPath: "assets/html/pages/password-reset-confirmation.tmpl", Cond: rsp.AccountsFull, Copy: true},

		{Template: "web/cmd/web/main.go.tmpl", OutPath: "cmd/web/main.go", Cond: true},
		{Template: "web/cmd/web/middleware.go.tmpl", OutPath: "cmd/web/middleware.go", Cond: true},
		{Template: "web/cmd/web/handlers.go.tmpl", OutPath: "cmd/web/handlers.go", Cond: true},
		{Template: "web/cmd/web/templates.go.tmpl", OutPath: "cmd/web/templates.go", Cond: true},
		{Template: "web/cmd/web/helpers.go.tmpl", OutPath: "cmd/web/helpers.go", Cond: true},
		{Template: "web/cmd/web/handlers-auth.go.tmpl", OutPath: "cmd/web/handlers-auth.go", Cond: rsp.AccountsFull},
		{Template: "web/cmd/web/context.go.tmpl", OutPath: "cmd/web/context.go", Cond: rsp.AccountsFull},
	}
}

// apiFiles returns the files for a JSON API.
func apiFiles(rsp spec.ResolvedSpec) []FileSpec {
	routes := fmt.Sprintf("api/cmd/api/routes/%s.go.tmpl", rsp.Router)
	return []FileSpec{
		{Template: routes, OutPath: "cmd/api/routes.go", Cond: true},
		{Template: "api/cmd/api/main.go.tmpl", OutPath: "cmd/api/main.go", Cond: true},
		{Template: "api/cmd/api/middleware.go.tmpl", OutPath: "cmd/api/middleware.go", Cond: true},
		{Template: "api/cmd/api/handlers.go.tmpl", OutPath: "cmd/api/handlers.go", Cond: true},
		{Template: "api/cmd/api/helpers.go.tmpl", OutPath: "cmd/api/helpers.go", Cond: true},
		{Template: "api/cmd/api/handlers-auth.go.tmpl", OutPath: "cmd/api/handlers-auth.go", Cond: rsp.AccountsJWT},
		{Template: "api/cmd/api/context.go.tmpl", OutPath: "cmd/api/context.go", Cond: rsp.AccountsJWT},
	}
}
