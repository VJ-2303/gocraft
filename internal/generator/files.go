package generator

import (
	"github.com/vj/nomoreboiler/internal/spec"
)

// FileSpec selects one template file for generation. Template is the path
// of the embedded template; OutPath is the destination path relative to the
// target directory (itself a template, so it can depend on the spec).
type FileSpec struct {
	Template string
	OutPath  string
	Cond     bool
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

// appFiles returns the application-type-specific files (web or api). The
// shared, web and api template trees land with the vertical-slice
// milestone; until then this returns an empty slice for the base app files.
func appFiles(rsp spec.ResolvedSpec) []FileSpec {
	_ = rsp
	return nil
}
