package controller

import (
	"btctl/dmc"
	"btctl/util"
	"strings"
	"time"
)

type Controller struct {
	user string
	binaryPath string

	state dmc.Decision

	decision dmc.Decision
	decisionChannel chan dmc.Decision

	controlTimer *time.Timer
	controllerExecuting bool
	controllerChannel chan dmc.Decision

	stop chan bool
	stopped chan bool
}

const controlPeriod time.Duration = time.Minute

var log = util.MustGetLogger("controller")

func New(user string, binaryPath string) *Controller {
	controller := &Controller{
		user: user,
		binaryPath: binaryPath,
		controllerChannel: make(chan dmc.Decision, 1),
		decisionChannel: make(chan dmc.Decision, 1),
		stop: make(chan bool, 1),
		stopped: make(chan bool),
	}
	go controller.loop()
	return controller
}

func (c *Controller) OnDecision(decision dmc.Decision) {
	c.decisionChannel <- decision
}

func (c *Controller) Close() {
	select {
	case c.stop <- true:
	}
	<- c.stopped
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
	command := []string{"sudo", "-H", "-u", c.user, c.binaryPath}

	if c.decision == dmc.START {
		command = append(command, "--start-all")
	} else if c.decision == dmc.STOP {
		command = append(command, "--stop-all")
	}

	commandString := strings.Join(command, " ")
	log.Debug("Executing `%s`...", commandString)

	go func(state dmc.Decision) {
		_, err := util.RunCmd(command[0], command[1:]...)

		if err != nil {
			log.Error("`%s` failed: %s", commandString, err)
			state = dmc.NO_DECISION
		}

		c.controllerChannel <- state
	}(c.decision)

	c.controlTimer.Stop()
	c.controllerExecuting = true
}

func (c *Controller) checkCurrentState() bool {
	if c.state != c.decision && !c.controllerExecuting {
		c.control()
		return false
	} else {
		return true
	}
}
