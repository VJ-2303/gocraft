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
		{Template: "shared/internal/env/env.go.tmpl", OutPath: "internal/env/env.go", Cond: rsp.ConfigSource == spec.ConfigEnv},
		{Template: "shared/internal/log/colorized.go.tmpl", OutPath: "internal/log/colorized.go", Cond: rsp.LogFormat == spec.LogColorized},

		// The assets package embeds migrations (both app types) and the
		// static files and HTML templates (web apps only).
		{Template: "shared/assets/efs.go.tmpl", OutPath: "assets/efs.go", Cond: rsp.AppType == spec.AppWeb || rsp.Migrations},
		{Template: "shared/assets/migrations/000001_init.up.sql", OutPath: "assets/migrations/000001_init.up.sql", Cond: rsp.Migrations, Copy: true},
		{Template: "shared/assets/migrations/000001_init.down.sql", OutPath: "assets/migrations/000001_init.down.sql", Cond: rsp.Migrations, Copy: true},
		{Template: "shared/assets/migrations/000002_create_sessions_table.up.sql.tmpl", OutPath: "assets/migrations/000002_create_sessions_table.up.sql", Cond: rsp.Sessions},
		{Template: "shared/assets/migrations/000002_create_sessions_table.down.sql.tmpl", OutPath: "assets/migrations/000002_create_sessions_table.down.sql", Cond: rsp.Sessions},

		{Template: "shared/internal/cookies/cookies.go.tmpl", OutPath: "internal/cookies/cookies.go", Cond: rsp.SecureCookies},
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
		{Template: "web/assets/html/components/nav.tmpl", OutPath: "assets/html/components/nav.tmpl", Cond: true, Copy: true},
		{Template: "web/assets/html/pages/home.tmpl", OutPath: "assets/html/pages/home.tmpl", Cond: true, Copy: true},

		{Template: "web/cmd/web/main.go.tmpl", OutPath: "cmd/web/main.go", Cond: true},
		{Template: "web/cmd/web/middleware.go.tmpl", OutPath: "cmd/web/middleware.go", Cond: true},
		{Template: "web/cmd/web/handlers.go.tmpl", OutPath: "cmd/web/handlers.go", Cond: true},
		{Template: "web/cmd/web/templates.go.tmpl", OutPath: "cmd/web/templates.go", Cond: true},
		{Template: "web/cmd/web/helpers.go.tmpl", OutPath: "cmd/web/helpers.go", Cond: true},
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
	}
}
