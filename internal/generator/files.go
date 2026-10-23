package generator

import (
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
	return []FileSpec{
		{Template: "shared/internal/response/html.go.tmpl", OutPath: "internal/response/html.go", Cond: true},

		{Template: "web/assets/efs.go.tmpl", OutPath: "assets/efs.go", Cond: true},
		{Template: "web/assets/static/css/main.css", OutPath: "assets/static/css/main.css", Cond: true, Copy: true},
		{Template: "web/assets/html/base.tmpl", OutPath: "assets/html/base.tmpl", Cond: true, Copy: true},
		{Template: "web/assets/html/components/nav.tmpl", OutPath: "assets/html/components/nav.tmpl", Cond: true, Copy: true},
		{Template: "web/assets/html/pages/home.tmpl", OutPath: "assets/html/pages/home.tmpl", Cond: true, Copy: true},

		{Template: "web/cmd/web/main.go.tmpl", OutPath: "cmd/web/main.go", Cond: true},
		{Template: "web/cmd/web/routes.go.tmpl", OutPath: "cmd/web/routes.go", Cond: true},
		{Template: "web/cmd/web/handlers.go.tmpl", OutPath: "cmd/web/handlers.go", Cond: true},
		{Template: "web/cmd/web/templates.go.tmpl", OutPath: "cmd/web/templates.go", Cond: true},
		{Template: "web/cmd/web/helpers.go.tmpl", OutPath: "cmd/web/helpers.go", Cond: true},
	}
}

// apiFiles returns the files for a JSON API.
func apiFiles(rsp spec.ResolvedSpec) []FileSpec {
	return []FileSpec{
		{Template: "api/cmd/api/main.go.tmpl", OutPath: "cmd/api/main.go", Cond: true},
		{Template: "api/cmd/api/routes.go.tmpl", OutPath: "cmd/api/routes.go", Cond: true},
		{Template: "api/cmd/api/handlers.go.tmpl", OutPath: "cmd/api/handlers.go", Cond: true},
		{Template: "api/cmd/api/helpers.go.tmpl", OutPath: "cmd/api/helpers.go", Cond: true},
	}
}
