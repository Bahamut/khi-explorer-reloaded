//go:build !windows

package tui

import (
	"os/exec"
	"syscall"
)

func pauseProcess(cmd *exec.Cmd) error {
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Signal(syscall.SIGSTOP)
	}
	return nil
}

func resumeProcess(cmd *exec.Cmd) error {
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Signal(syscall.SIGCONT)
	}
	return nil
}

func killProcess(cmd *exec.Cmd) error {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGCONT)
		return cmd.Process.Kill()
	}
	return nil
}

func applyPlatformWindowSettings() {}
