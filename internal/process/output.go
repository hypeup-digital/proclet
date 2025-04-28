package process

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/hypeup-digital/proclet/config"
	"github.com/hypeup-digital/proclet/internal/fail"
	"github.com/hypeup-digital/proclet/internal/input"
	"github.com/pkg/term/termios"

	"github.com/fatih/color"
)

var timestampColor = color.New(color.FgWhite)

type ptyPipe struct {
	pty, tty *os.File
}

type MultiOutput struct {
	maxNameLength  int
	mutex          sync.Mutex
	pipes          map[*Process]*ptyPipe
	printProcName  bool
	printTimestamp bool
}

func NewMultiOutput(config config.Config) *MultiOutput {
	return &MultiOutput{
		printProcName:  config.Output.PrintAppNames,
		printTimestamp: config.Output.PrintTimeStamps,
		maxNameLength:  config.Output.MaxAppNameLength,
	}
}

func (m *MultiOutput) openPipe(proc *Process) (pipe *ptyPipe) {
	var err error

	pipe = m.pipes[proc]

	pipe.pty, pipe.tty, err = termios.Pty()
	fail.FatalOnErr(err)

	proc.Stdout = pipe.tty
	proc.Stderr = pipe.tty
	proc.Stdin = pipe.tty
	proc.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}

	return
}

func (m *MultiOutput) Connect(proc *Process) {
	if len(proc.Name) > m.maxNameLength {
		m.maxNameLength = len(proc.Name)
	}

	if m.pipes == nil {
		m.pipes = make(map[*Process]*ptyPipe)
	}

	m.pipes[proc] = &ptyPipe{}
}

func (m *MultiOutput) PipeOutput(proc *Process) {
	pipe := m.openPipe(proc)

	go func(proc *Process, pipe *ptyPipe) {
		input.ScanLines(pipe.pty, func(b []byte) bool {
			m.WriteLine(proc, b)
			return true
		})
	}(proc, pipe)
}

func (m *MultiOutput) ClosePipe(proc *Process) {
	if pipe := m.pipes[proc]; pipe != nil {
		pipe.pty.Close()
		pipe.tty.Close()
	}
}

func (m *MultiOutput) WriteLine(proc *Process, p []byte) {
	var buf bytes.Buffer

	if m.printProcName || m.printTimestamp {
		if m.printTimestamp {
			timestampColor.Fprintf(&buf, "%s ", time.Now().Format("15:04:05.000"))
			buf.WriteByte(' ')
		}

		if m.printProcName {
			proc.Color.Fprintf(&buf, "| %-*s | ", m.maxNameLength, proc.Name)
		}
	}

	buf.WriteByte(' ')
	buf.Write(p)
	buf.WriteByte('\n')

	m.mutex.Lock()
	defer m.mutex.Unlock()

	buf.WriteTo(os.Stdout)
}

func (m *MultiOutput) WriteErr(proc *Process, err error) {
	m.WriteLine(proc, []byte(
		fmt.Sprintf("\033[0;31m%v\033[0m", err),
	))
}
