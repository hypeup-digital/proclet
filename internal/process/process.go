package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/fatih/color"
)

type Process struct {
	*exec.Cmd

	Name  string
	Color *color.Color

	output *MultiOutput
}

func NewProcess(name, command string, color *color.Color, output *MultiOutput) *Process {
	proc := &Process{
		Cmd:    exec.Command("/bin/sh", "-c", command),
		Name:   name,
		Color:  color,
		output: output,
	}

	proc.output.Connect(proc)
	return proc
}

func (p *Process) writeLine(b []byte) {
	p.output.WriteLine(p, b)
}

func (p *Process) writeErr(err error) {
	p.output.WriteErr(p, err)
}

func (p *Process) signal(sig os.Signal) {
	group, err := os.FindProcess(-p.Process.Pid)
	if err != nil {
		p.writeErr(err)
		return
	}

	if err = group.Signal(sig); err != nil {
		p.writeErr(err)
	}
}

func (p *Process) Running() bool {
	return p.Process != nil && p.ProcessState == nil
}

func (p *Process) Run() {
	p.output.PipeOutput(p)
	defer p.output.ClosePipe(p)

	ensureKill(p)

	p.writeLine([]byte("\033[1mProcess started\033[0m"))

	if err := p.Cmd.Run(); err != nil {
		p.writeErr(err)
	} else {
		p.writeLine([]byte(fmt.Sprintf("\033[1mProcess exited with code %d\033[0m", p.Cmd.ProcessState.ExitCode())))
	}
}

func (p *Process) Interrupt() {
	if p.Running() {
		p.writeLine([]byte("\033[1mInterrupting process\033[0m"))
		p.signal(syscall.SIGINT)
	}
}

func (p *Process) Kill() {
	if p.Running() {
		p.writeLine([]byte("\033[1mKilling process\033[0m"))
		p.signal(syscall.SIGKILL)
	}
}
