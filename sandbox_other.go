//go:build !darwin && !linux

package main

import (
	"context"
	"fmt"
	"os/exec"
)

func checkSandbox() error {
	return fmt.Errorf("no sandbox available on this platform")
}

func sandboxedCommand(shell, text string) *exec.Cmd {
	return exec.Command(shell, "-c", text)
}

func sandboxedCommandContext(ctx context.Context, shell, text string) *exec.Cmd {
	return exec.CommandContext(ctx, shell, "-c", text)
}

func runInSandbox() {}
