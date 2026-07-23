// git-wt provides safe, project-oriented Git worktree commands through `git wt`.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jamesonstone/kit/internal/worktree"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git wt: determine current directory: %v\n", err)
		os.Exit(1)
	}

	app := worktree.NewApp(os.Stdout, os.Stderr)
	if err := app.Run(context.Background(), cwd, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "git wt: %v\n", err)
		os.Exit(1)
	}
}
