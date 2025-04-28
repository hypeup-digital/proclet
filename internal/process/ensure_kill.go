//go:build !linux

package process

func ensureKill(p *Process) {
	// p.SysProcAttr.Pdeathsig is supported only on Linux, we can't do anything here
}
