// +build linux

package ps

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// UnixProcess is an implementation of Process that contains Unix-specific
// fields and information.
type UnixProcess struct {
	pid    int
	ppid   int
	state  rune
	pgrp   int
	sid    int
	cpids  []int
	binary string
}

func (p *UnixProcess) Pid() int {
	return p.pid
}

func (p *UnixProcess) PPid() int {
	return p.ppid
}

func (p *UnixProcess) CPids() []int {
	return p.cpids
}

func (p *UnixProcess) Executable() string {
	return p.binary
}

// Refresh reloads all the data associated with this process.
func (p *UnixProcess) Refresh() error {
	statPath := fmt.Sprintf("/proc/%d/stat", p.pid)
	dataBytes, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}

	// stat file have truncated names
	// First, parse out the image name
	data := string(dataBytes)
	binStart := strings.IndexRune(data, '(') + 1
	binEnd := strings.IndexRune(data[binStart:], ')')

	for _, str := range []string{"comm", "cmdline"} {
		dataBytes, err = ioutil.ReadFile(fmt.Sprintf("/proc/%d/%s", p.pid, str))
		if err != nil {
			continue
		}

		if p.binary = strings.TrimSpace(string(bytes.Trim(dataBytes, "\x00"))); p.binary != "" {
			break
		}
	}

	if p.binary == "" {
		return fmt.Errorf("failed to get process executable")
	}

	// Move past the image name and start parsing the rest
	data = data[binStart+binEnd+2:]
	_, err = fmt.Sscanf(data,
		"%c %d %d %d",
		&p.state,
		&p.ppid,
		&p.pgrp,
		&p.sid)

	f, err := os.Open(fmt.Sprintf("/proc/%d/task", p.pid))
	if err != nil {
		return err
	}
	defer f.Close()

	dirnames, err := f.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range dirnames {
		pid, err := strconv.ParseInt(name, 10, 0)
		if err != nil {
			continue
		}
		p.cpids = append(p.cpids, int(pid))
	}

	return err
}

func findProcessByPid(pid int) (Process, error) {
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

func findProcessByExecutable(name string) ([]Process, error) {
	var p []Process
	procs, err := processes()
	if err != nil {
		return nil, err
	}
	for _, proc := range procs {
		if strings.Compare(proc.Executable(), name) == 0 {
			p = append(p, proc)
		}
	}
	return p, nil
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
