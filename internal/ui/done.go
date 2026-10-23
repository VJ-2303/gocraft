package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/vj/nomoreboiler/internal/generator"
	"github.com/vj/nomoreboiler/internal/spec"
)

// doneStyles holds the styling for the final summary screen.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	fieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	valueStyle = lipgloss.NewStyle().
			Bold(true)
)

// printDone shows the final summary and tailored getting-started
// instructions, mirroring the confirmation page of the product this tool
// replicates.
func printDone(rsp spec.ResolvedSpec, result generator.Result) {
	fmt.Println()
	fmt.Println(titleStyle.Render("✔ Your codebase is ready!"))
	fmt.Println()

	summary := []struct{ label, value string }{
		{"Application type", string(rsp.AppType)},
		{"Module path", rsp.ModulePath},
		{"Database", string(rsp.Dialect)},
		{"Router", string(rsp.Router)},
		{"Config source", string(rsp.ConfigSource)},
		{"Log format", string(rsp.LogFormat)},
		{"Files generated", fmt.Sprintf("%d", len(result.Files))},
		{"Generated in", result.Dir},
	}
	for _, row := range summary {
		fmt.Printf("  %s %s\n",
			fieldStyle.Render(pad(row.label, 18)),
			valueStyle.Render(row.value))
	}

	fmt.Println()
	fmt.Println(titleStyle.Render("Getting started"))
	fmt.Println()
	fmt.Printf("  cd %s\n", result.Dir)
	fmt.Printf("  make run\n")

	var extra []string
	if rsp.Migrations {
		extra = append(extra,
			fmt.Sprintf("  make migrations/up   # apply the SQL migrations (DSN: %s)", rsp.Driver.ExampleDSN))
	}
	if rsp.ConfigSource == spec.ConfigFlags {
		extra = append(extra, "  make run -- --help   # show every configuration flag and default")
	} else {
		extra = append(extra, "  Configuration settings are read from environment variables; see cmd/"+rsp.CmdDir+"/main.go.")
	}
	if len(extra) > 0 {
		fmt.Println()
		fmt.Println(strings.Join(extra, "\n"))
	}

	fmt.Println()
	fmt.Println("Everything in the codebase is designed to be edited — it's your code.")
	fmt.Println("Run make help inside the project to see the available admin commands.")
}

// pad right-pads s with spaces to the given width.
func pad(s string, width int) string {
	for len(s) < width {
		s += " "
	}
	return s
}
