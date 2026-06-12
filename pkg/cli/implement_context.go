package cli

import (
	"fmt"

	"github.com/jamesonstone/kit/internal/document"
	"github.com/jamesonstone/kit/internal/feature"
)

func printImplementationContext(feat *feature.Feature, brainstormPath, specPath, planPath, tasksPath, summary string, progress feature.TaskProgress) {
	hasBrainstorm := document.Exists(brainstormPath)
	style := styleForStdout()

	fmt.Println()
	fmt.Println(style.title("🛠️", fmt.Sprintf("Implementation Context: %s", feat.DirName)))
	fmt.Println()

	if summary != "" {
		fmt.Println(style.title("📝", "Feature Summary"))
		fmt.Println(summary)
		fmt.Println()
	} else {
		fmt.Println(style.title("📝", "Feature Summary"))
		fmt.Println("(Read SPEC.md for feature description)")
		fmt.Println()
	}

	if progress.HasTasks() {
		fmt.Println(style.title("📈", fmt.Sprintf("Progress: %d/%d tasks complete", progress.Complete, progress.Total)))
	} else {
		fmt.Println(style.title("📈", "Progress: Tasks defined, ready to begin"))
	}
	fmt.Println()

	fmt.Println(style.title("📚", "Document Reference"))
	printImplementDocumentReferenceTable()
	fmt.Println()

	fmt.Println(style.title("📍", "File Locations"))
	if hasBrainstorm {
		fmt.Printf("  • BRAINSTORM: %s\n", brainstormPath)
	}
	fmt.Printf("  • SPEC:  %s\n", specPath)
	fmt.Printf("  • PLAN:  %s\n", planPath)
	fmt.Printf("  • TASKS: %s\n", tasksPath)
	fmt.Println()
}
