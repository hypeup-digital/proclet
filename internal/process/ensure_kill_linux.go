//go:build linux

package process

import "syscall"

func ensureKill(p *Process) {
	p.SysProcAttr.Pdeathsig = syscall.SIGKILL
}
