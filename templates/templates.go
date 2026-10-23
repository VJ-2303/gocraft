// Package templates embeds the file templates used to generate codebases.
//
// Layout:
//
//	templates/common/   root files shared by every app type (go.mod, Makefile, ...)
//	templates/shared/   internal/ packages shared between web apps and APIs
//	templates/web/      files specific to traditional web applications
//	templates/api/      files specific to JSON APIs
//
// The shared, web and api trees are added as they land.
package templates

import "embed"

// FS holds every template file. Templates are rendered by the generator
// package with a spec.ResolvedSpec as their data.
//
//go:embed common
var FS embed.FS
