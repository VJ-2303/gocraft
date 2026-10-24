package ui

import (
	"fmt"

	"github.com/VJ-2303/gocraft/internal/generator"
	"github.com/VJ-2303/gocraft/internal/spec"
)

// printDone shows the final summary and getting-started instructions.
func printDone(rsp spec.ResolvedSpec, result generator.Result) {
	fmt.Println()
	fmt.Println(successStyle.Render("✔ Your codebase is ready!"))
	fmt.Println()

	summary := []struct{ label, value string }{
		{"Application type", string(rsp.AppType)},
		{"Module path", rsp.ModulePath},
		{"Database", string(rsp.Dialect)},
		{"Router", string(rsp.Router)},
		{"Files generated", fmt.Sprintf("%d", len(result.Files))},
		{"Generated in", result.Dir},
	}
	for _, row := range summary {
		fmt.Printf("  %-18s %s\n", row.label, row.value)
	}

	fmt.Println()
	fmt.Println(successStyle.Render("Getting started"))
	fmt.Println()

	commands := []string{result.Dir, "make run"}
	if rsp.Migrations {
		commands = append(commands, "make migrations/up")
	}
	for _, command := range commands {
		fmt.Println(commandStyle.Render("  $ " + command))
	}

	fmt.Println()
	fmt.Println(hintStyle.Render("Everything in the codebase is designed to be edited — it's your code."))
	fmt.Println(hintStyle.Render("Run make help inside the project to see the available admin commands."))
}

// printCanceled explains the state after the user cancels generation.
func printCanceled(target string) {
	fmt.Println(warningStyle.Render("✗ Generation canceled."))
	if target != "" {
		fmt.Printf("The target directory %s may be partially populated; remove it before retrying.\n", target)
	}
}

// printAborted confirms a clean exit after the user aborts the wizard.
func printAborted() {
	fmt.Println(hintStyle.Render("Aborted. No codebase was generated."))
}

// printFailed shows a failure screen with the error and guidance on the
// target directory state.
func printFailed(err error) {
	fmt.Println(warningStyle.Render("✗ Generation failed."))
	fmt.Println()
	fmt.Printf("  %v\n", err)
	fmt.Println()
	fmt.Println(hintStyle.Render("Files already written remain in the target directory."))
	fmt.Println(hintStyle.Render("Fix the issue, remove the directory if you want a clean start,"))
	fmt.Println(hintStyle.Render("and run the wizard again."))
}
