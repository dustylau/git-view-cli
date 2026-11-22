package main

import (
	"context"
	"os"
	"os/signal"

	"git-view-cli/internal/cli"
	gitview "git-view-cli/internal/git"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr, func(ctx context.Context, directory string) (gitview.Snapshot, error) {
		return gitview.Read(ctx, gitview.Process{}, directory)
	})
	stop()
	os.Exit(code)
}
