package util

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type MainLoop interface {
	Init() error
	TickHandler() error
	Close()
}

func Error(message string, args ...interface{}) error {
	if len(args) == 0 {
		return errors.New(message)
	} else {
		return errors.New(fmt.Sprintf(message, args...))
	}
}

func SetupSignalHandling() <-chan os.Signal {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return channel
}

func Loop(mainLoop MainLoop, tickInterval time.Duration) (err error) {
	signalChannel := SetupSignalHandling()

	defer mainLoop.Close()
	err = mainLoop.Init()
	if err != nil {
		return
	}

	tickTimer := time.NewTimer(tickInterval)
	defer tickTimer.Stop()

	for stop := false; !stop; {
		tickTimer.Reset(tickInterval)

		select {
		case <-signalChannel:
			log.Println("Got termination signal. Exiting...")
			stop = true
		case <-tickTimer.C:
			err = mainLoop.TickHandler()
			if err != nil {
				log.Println(err)
				stop = true
			}
		}
	}

	if err != nil {
		os.Exit(1)
	}

	return
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
