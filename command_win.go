//go:build windows

package termview

import (
	"os"
	"os/exec"
)

func getShellPath() string {
	command := os.Getenv("COMSPEC")
	if command == "" {
		command = `C:\Windows\System32\cmd.exe`
	}
	return command
}

func buildCommand(cmd string) *exec.Cmd {
	cmdExec := exec.Command(cmd)
	cmdExec.Env = append(os.Environ(), "TERM=xterm-256color")
	return cmdExec
}
