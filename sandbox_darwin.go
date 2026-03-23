//go:build darwin

package main

import (
	"context"
	"fmt"
	"os/exec"
)

// seatbeltProfile is a read-only Apple Seatbelt (sandbox-exec) profile.
// It allows reading any file and executing processes, but denies all
// file writes and other sensitive operations.
// Inherited pipe fds (stdout/stderr) are unaffected because Seatbelt
// file-write rules gate open(path, O_WRONLY) calls, not write(fd) on
// already-open descriptors.
const seatbeltProfile = `(version 1)
(deny default)
(allow file-read*)
(allow process-exec*)
(allow process-fork)
(allow signal)
(allow sysctl-read)
(allow mach*)`

func checkSandbox() error {
	_, err := exec.LookPath("sandbox-exec")
	if err != nil {
		return fmt.Errorf("sandbox-exec not found: %w", err)
	}
	return nil
}

func sandboxedCommand(shell, text string) *exec.Cmd {
	return exec.Command("sandbox-exec", "-p", seatbeltProfile, shell, "-c", text)
}

func sandboxedCommandContext(ctx context.Context, shell, text string) *exec.Cmd {
	return exec.CommandContext(ctx, "sandbox-exec", "-p", seatbeltProfile, shell, "-c", text)
}

// runInSandbox is a no-op on Darwin; sandboxing is applied per-command
// via sandbox-exec rather than through self-re-execution.
func runInSandbox() {}
