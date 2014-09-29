package util

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

func RunCmd(name string, args ...string) (stdout string, runError error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			error := strings.TrimLeft(stderrBuf.String(), "\n")
			error = strings.Split(error, "\n")[0]
			return "", errors.New(error)
		} else {
			return "", err
		}
	}

	return stdoutBuf.String(), nil
}
