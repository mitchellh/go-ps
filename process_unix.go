// +build linux

package ps

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

// UnixProcess is an implementation of Process that contains Unix-specific
// fields and information.
type UnixProcess struct {
	pid  int
	ppid int
	//state rune
	//pgrp  int
	//sid   int

	binary string
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

// Refresh reloads all the data associated with this process.
func (p *UnixProcess) Refresh() error {
	var (
		procPath = fmt.Sprintf("/proc/%d/", p.pid)
		data     []byte
		err      error
	)

	data, err = ioutil.ReadFile(procPath + "stat")
	if err != nil {
		return err
	}

	data = data[bytes.IndexRune(data, ')')+2:]
	_, err = fmt.Fscanf(bytes.NewReader(data), "%c %d", new(rune), &p.ppid)
	if err != nil {
		return err
	}

	data, err = ioutil.ReadFile(procPath + "cmdline")
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("empty name")
	}

	// Remove arguments
	if ind := bytes.IndexByte(data, 0); ind >= 0 {
		data = data[:ind]
	}
	if ind := bytes.IndexByte(data, ' '); ind >= 0 {
		data = data[:ind]
	}
	// Remove path to the executable
	if ind := bytes.LastIndexByte(data, '/'); ind >= 0 {
		data = data[ind+1:]
	}

	p.binary = string(data)

	return nil
}

func findProcess(pid int) (Process, error) {
	dir := fmt.Sprintf("/proc/%d", pid)
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
	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	results := make([]Process, 0, 50)
	for {
		fis, err := d.Readdir(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, fi := range fis {
			// We only care about directories, since all pids are dirs
			if !fi.IsDir() {
				continue
			}

			// We only care if the name starts with a numeric
			name := fi.Name()
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
