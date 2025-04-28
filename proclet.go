package proclet

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hypeup-digital/proclet/config"
	"github.com/hypeup-digital/proclet/internal/process"
)

type proclet struct {
	config config.Config

	output *process.MultiOutput

	processes []*process.Process

	waitGroup           sync.WaitGroup
	doneChannel         chan bool
	interruptionChannel chan os.Signal
}

type Proclet interface {
	Run()
}

func FromConfig(conf config.Config) (Proclet, error) {
	proc := proclet{
		config: conf,
	}

	// Initialize output for managed processes
	proc.output = process.NewMultiOutput(conf)

	// Register all processes
	for i, application := range conf.Applications {
		processColor := config.AppColors[i%len(config.AppColors)]
		proc.processes = append(proc.processes, process.NewProcess(application.Identifier, application.Command, processColor, proc.output))
	}

	return &proc, nil
}

func (p *proclet) Run() {
	// Print banner if it was provided
	if len(p.config.Banner) > 0 {
		fmt.Printf("%s\n\n", p.config.Banner)
	}

	// Initialize channels
	p.doneChannel = make(chan bool, len(p.processes))
	p.interruptionChannel = make(chan os.Signal)

	// Wire up signals to interruption channel
	signal.Notify(p.interruptionChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start all processes
	for _, proc := range p.processes {
		p.runProcess(proc)
	}

	go p.waitForExit()

	p.waitGroup.Wait()
}

func (p *proclet) runProcess(proc *process.Process) {
	p.waitGroup.Add(1)

	go func() {
		defer p.waitGroup.Done()
		defer func() { p.doneChannel <- true }()

		proc.Run()
	}()
}

func (p *proclet) waitForDoneOrInterrupt() {
	select {
	case <-p.doneChannel:
	case <-p.interruptionChannel:
	}
}

func (p *proclet) waitForTimeoutOrInterrupt() {
	select {
	case <-time.After(p.config.Timeout):
	case <-p.interruptionChannel:
	}
}

func (p *proclet) waitForExit() {
	p.waitForDoneOrInterrupt()

	for _, proc := range p.processes {
		go proc.Interrupt()
	}

	p.waitForTimeoutOrInterrupt()

	for _, proc := range p.processes {
		go proc.Kill()
	}
}
