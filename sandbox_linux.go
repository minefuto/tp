//go:build linux

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/landlock-lsm/go-landlock/landlock"
	llsyscall "github.com/landlock-lsm/go-landlock/landlock/syscall"
)

const sandboxEnvVar = "TP_SANDBOX_EXEC"

var selfExe string // set by checkSandbox()

// checkSandbox verifies that Landlock V3+ is supported by the running kernel
// and that the tp executable path is resolvable for self-re-execution.
func checkSandbox() error {
	abi, err := llsyscall.LandlockGetABIVersion()
	if err != nil {
		return fmt.Errorf("Landlock is not supported by this kernel: %w", err)
	}
	if abi < 3 {
		return fmt.Errorf("Landlock ABI v%d is too old (need v3+)", abi)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	selfExe = exe
	return nil
}

// runInSandbox checks if the current process is the re-executed sandbox
// worker. If so, it applies Landlock, then execve's the shell command
// encoded in the process arguments, replacing the process image.
//
// This must be called at the very start of main(), before flag.Parse(),
// because it inspects os.Args directly.
func runInSandbox() {
	if os.Getenv(sandboxEnvVar) != "1" {
		return
	}

	// Expected args: tp <shell> -c <text>
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "tp sandbox: invalid args: %v\n", os.Args)
		os.Exit(1)
	}

	if err := landlock.V3.RestrictPaths(landlock.RODirs("/")); err != nil {
		fmt.Fprintf(os.Stderr, "tp sandbox: landlock: %v\n", err)
		os.Exit(1)
	}

	// Replace this process image with the shell. Strip TP_SANDBOX_EXEC
	// from the environment so recursive invocations of tp don't enter
	// sandbox-worker mode.
	if err := syscall.Exec(os.Args[1], os.Args[1:], filteredEnv()); err != nil {
		fmt.Fprintf(os.Stderr, "tp sandbox: execve(%q): %v\n", os.Args[1], err)
		os.Exit(1)
	}
}

// filteredEnv returns os.Environ() with TP_SANDBOX_EXEC removed.
func filteredEnv() []string {
	env := os.Environ()
	result := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, sandboxEnvVar+"=") {
			result = append(result, e)
		}
	}
	return result
}

func sandboxedCommand(shell, text string) *exec.Cmd {
	if selfExe == "" {
		return exec.Command(shell, "-c", text)
	}
	cmd := exec.Command(selfExe, shell, "-c", text)
	cmd.Env = append(filteredEnv(), sandboxEnvVar+"=1")
	return cmd
}

func sandboxedCommandContext(ctx context.Context, shell, text string) *exec.Cmd {
	if selfExe == "" {
		return exec.CommandContext(ctx, shell, "-c", text)
	}
	cmd := exec.CommandContext(ctx, selfExe, shell, "-c", text)
	cmd.Env = append(filteredEnv(), sandboxEnvVar+"=1")
	return cmd
}
