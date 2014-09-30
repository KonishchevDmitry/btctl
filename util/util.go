package util

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func Error(message string, args ...interface{}) error {
	if len(args) == 0 {
		return errors.New(message)
	} else {
		return errors.New(fmt.Sprintf(message, args...))
	}
}

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
			return "", Error(error)
		} else {
			return "", err
		}
	}

	return stdoutBuf.String(), nil
}
