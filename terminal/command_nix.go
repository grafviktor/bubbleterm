//go:build !windows

package terminal

import (
	"os"
	"os/exec"
	"syscall"
)

func getShellPath() string {
	command := os.Getenv("SHELL")
	if command == "" {
		command = "/bin/sh"
	}

	return command
}

func buildCommand(cmd string) *exec.Cmd {
	cmdExec := exec.Command(cmd)
	cmdExec.Env = append(os.Environ(), "TERM=xterm-256color")
	// Should match creack/pty.StartWithSize behavior. See here:
	// https://github.com/creack/pty/blob/v1.1.24/start.go#L18-L24
	// Without those attrs, SIGWINCH is not delivered to the shell on resize
	// and resize does not work properly.
	cmdExec.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	return cmdExec
}
