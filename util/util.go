package util

import (
	"btctl/logging"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type mainLoop interface {
	Init() error
	OnTick() error
	Close()
}

var log = MustGetLogger("util")
var debugMode = false;

func InitFlags() {
	flag.BoolVar(&debugMode, "debug", false, "debug mode")
	flag.Parse()
}

func MustInitLogging(withTime bool) {
	level := logging.INFO
	format := "%{level:.1s}: %{message}"
	timeFormat := ""

	if withTime {
		timeFormat = "%{time:2006.01.02 15:04:05} "
	}

	if debugMode {
		level = logging.DEBUG
		format = " %{shortfile} " + format
		timeFormat = "%{time:2006.01.02 15:04:05.000} "
	}

	format = timeFormat + format

	logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))
	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetLevel(level, "")
}

func MustGetLogger(name string) *logging.Logger {
	return logging.MustGetLogger(name)
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

func Loop(mainLoop mainLoop, tickInterval time.Duration) {
	var err error = nil

	signalChannel := SetupSignalHandling()

	defer func() {
		if err != nil {
			log.Error("%s", err)
			os.Exit(1)
		}
	}()

	defer mainLoop.Close()
	err = mainLoop.Init()
	if err != nil {
		return
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	loop:
		for {
			err = mainLoop.OnTick()
			if err != nil {
				break
			}

			select {
			case <-signalChannel:
				log.Info("Got a termination signal. Exiting...")
				break loop
			case <-ticker.C:
			}
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
