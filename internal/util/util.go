package util

import (
	"os"
	"os/exec"
)

func ExecuteCommand(dir string, cmd string, args... string) error {
	osCmd := exec.Command(cmd, args...)
	osCmd.Dir = dir
	osCmd.Stderr = os.Stderr

	return osCmd.Run()
}

