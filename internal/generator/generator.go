// Package generator renders the embedded file templates for a resolved
// spec and writes a complete codebase to a target directory.
package generator

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vj/nomoreboiler/internal/spec"
	"github.com/vj/nomoreboiler/templates"
)

// Result describes a completed generation run.
type Result struct {
	// Dir is the absolute path of the generated codebase.
	Dir string

	// Files lists the generated files, relative to Dir, in write order.
	Files []string
}

// Generate writes a codebase for the resolved spec into targetDir. The
// directory must not exist, or must exist and be empty. Files are written
// with 0644 permissions and directories with 0755.
func Generate(rsp spec.ResolvedSpec, targetDir string, logger *slog.Logger) (Result, error) {
	if err := rsp.Validate(); err != nil {
		return Result{}, fmt.Errorf("invalid codebase spec: %w", err)
	}

	if err := checkTarget(targetDir); err != nil {
		return Result{}, fmt.Errorf("checking target directory: %w", err)
	}

	result := Result{Dir: targetDir}

	createdTarget := false
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		createdTarget = true
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("creating target directory: %w", err)
	}
	// If generation fails before writing any file and we created the target
	// directory ourselves, remove it again so a retry starts clean.
	defer func() {
		if createdTarget && len(result.Files) == 0 {
			os.RemoveAll(targetDir)
		}
	}()

	for _, f := range fileMatrix(rsp) {
		out, err := renderTemplate("outpath:"+f.Template, f.OutPath, rsp)
		if err != nil {
			return result, fmt.Errorf("rendering output path for %s: %w", f.Template, err)
		}
		out = filepath.FromSlash(out)

		content, err := renderTemplate(f.Template, readFile(f.Template), rsp)
		if err != nil {
			return result, fmt.Errorf("rendering %s: %w", f.Template, err)
		}

		dest := filepath.Join(targetDir, out)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return result, fmt.Errorf("creating directory for %s: %w", out, err)
		}
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			return result, fmt.Errorf("writing %s: %w", out, err)
		}

		result.Files = append(result.Files, out)
		logger.Debug("generated file", "path", out)
	}

	return result, nil
}

// checkTarget reports an error if the target directory exists and contains
// anything.
func checkTarget(targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("%s is not empty", targetDir)
	}
	return nil
}

// readFile returns the contents of a template from the embedded FS.
func readFile(name string) string {
	data, err := templates.FS.ReadFile(name)
	if err != nil {
		// Template paths are fixed at compile time by the file matrix; a
		// miss is a programming error, not a runtime condition.
		panic(fmt.Sprintf("template %s not found in embedded FS: %v", name, err))
	}
	return string(data)
}

// renderTemplate parses and executes template text with the resolved spec
// as its data. Unknown keys are an error, so typos in templates fail loudly
// at generation time.
func renderTemplate(name, text string, rsp spec.ResolvedSpec) (string, error) {
	tmpl, err := template.New(name).Funcs(funcMap).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", fmt.Errorf("parsing: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, rsp); err != nil {
		return "", fmt.Errorf("executing: %w", err)
	}
	return sb.String(), nil
}

// funcMap holds the custom template functions available to file templates.
var funcMap = template.FuncMap{
	// indent prefixes every line of s with the given string, so generated
	// code blocks can be nested inside conditionally-emitted wrappers.
	"indent": func(prefix, s string) string {
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			if line != "" {
				lines[i] = prefix + line
			}
		}
		return strings.Join(lines, "\n")
	},
}
