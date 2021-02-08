// +build linux solaris

package ps

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

// UnixProcess is an implementation of Process that contains Unix-specific
// fields and information.
type UnixProcess struct {
	pid   int
	ppid  int
	state rune
	pgrp  int
	sid   int

	binary  string
	cmdline string
}

const (
	UnixProcRoot = "/proc"
)

func procDir(pid int) string {
	return path.Join(UnixProcRoot, fmt.Sprintf("%v", pid))
}

func (p *UnixProcess) Pid() int {
	return p.pid
}

func (p *UnixProcess) PPid() int {
	return p.ppid
}

func (p *UnixProcess) Executable() string {
	return p.binary
}

// CMDLine Full command of process
func (p *UnixProcess) CMDLine() string {
	return p.cmdline
}

func findProcess(pid int) (Process, error) {
	dir := procDir(pid)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	return newUnixProcess(pid)
}

func processes() ([]Process, error) {
	d, err := os.Open(UnixProcRoot)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	results := make([]Process, 0, 50)
	for {
		names, err := d.Readdirnames(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, name := range names {
			// We only care if the name starts with a numeric
			if name[0] < '0' || name[0] > '9' {
				continue
			}

			// From this point forward, any errors we just ignore, because
			// it might simply be that the process doesn't exist anymore.
			pid, err := strconv.ParseInt(name, 10, 0)
			if err != nil {
				continue
			}

			p, err := newUnixProcess(int(pid))
			if err != nil {
				continue
			}

			results = append(results, p)
		}
	}

	return results, nil
}

func newUnixProcess(pid int) (*UnixProcess, error) {
	p := &UnixProcess{pid: pid}
	return p, p.Refresh()
}
