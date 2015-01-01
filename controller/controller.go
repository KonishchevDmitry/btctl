package controller

import (
	"btctl/dmc"
	"btctl/util"
	"strings"
	"time"
)

type Controller struct {
	state dmc.Decision
	decision dmc.Decision
	controlTimer *time.Timer
	controllerExecuting bool
	controllerChannel chan dmc.Decision
	decisionChannel chan dmc.Decision
	stopped chan bool
	stop chan bool
}

const controlPeriod time.Duration = 5 * time.Second

var log = util.MustGetLogger("controller")

func (c *Controller) Init() {
	// TODO
	c.controllerChannel = make(chan dmc.Decision)
	c.decisionChannel = make(chan dmc.Decision, 1)
	c.stopped = make(chan bool)
	c.stop = make(chan bool, 1)
	go c.loop()
}

func (c *Controller) OnDecision(decision dmc.Decision) {
	c.decisionChannel <- decision
}

func (c *Controller) Close() {
	if c.stop != nil {
		select {
		case c.stop <- true:
		}
		<- c.stopped
	}
}

func (c *Controller) loop() {
	defer close(c.stopped)

	c.controlTimer = time.NewTimer(0)
	defer c.controlTimer.Stop()

	loop:
		for {
			select {
			case <-c.stop:
				if c.controllerExecuting {
					<- c.controllerChannel
				}
				break loop
			case <-c.controlTimer.C:
				c.control()
			case c.state = <-c.controllerChannel:
				c.controllerExecuting = false
				if c.state == dmc.NO_DECISION || c.checkCurrentState() {
					c.controlTimer.Reset(controlPeriod)
				}
			case decision := <-c.decisionChannel:
				if decision != dmc.NO_DECISION && c.decision != decision {
					log.Debug("Got new decision: %s", decision)
					c.decision = decision
					c.checkCurrentState()
				}
			}
		}
}

func (c *Controller) control() {
	var action string

	if c.decision == dmc.START {
		action = "start-all"
	} else if c.decision == dmc.STOP {
		action = "stop-all"
	}

	c.controlTimer.Stop()
	c.controllerExecuting = true

	go func(decision dmc.Decision) {
		command := make([]string, 0)
		command = append(command, "ssh", "server.lan", "/opt/bin/transmission-controller")
		if action != "" {
			command = append(command, "--" + action)
		}
		log.Debug("Executing `%s`...", strings.Join(command, " "))
		//		command := "false"
		//		stdout, err := util.RunCmd(command, "GGGG")
		//
		//		if err == nil {
		//			log.Error("%s", stdout)
		//		} else {
		//			err = util.Error("%s execution error: %s", command, err)
		//			log.Error("%s", err)
		//			decision = dmc.NO_DECISION
		//		}

		c.controllerChannel <- decision
	}(c.decision)
}

func (c *Controller) checkCurrentState() bool {
	if c.state != c.decision && !c.controllerExecuting {
		c.control()
		return false
	} else {
		return true
	}
}
