// +build plan9

package ps

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// Plan9Process is an implementation of Process that contains Plan9-specific
// fields and information.
type Plan9Process struct {
	pid    int
	ppid   int
	binary string
}

func (p *Plan9Process) Pid() int {
	return p.pid
}

func (p *Plan9Process) PPid() int {
	return p.ppid
}

func (p *Plan9Process) Executable() string {
	return p.binary
}

func (p *Plan9Process) Refresh() error {
	statusPath := fmt.Sprintf("/proc/%d/status", p.pid)
	statusDataBytes, err := ioutil.ReadFile(statusPath)
	if err != nil {
		return err
	}

	// Parse out name of binary
	data := string(statusDataBytes)
	binStart := 0
	binEnd := strings.IndexRune(data[binStart:], ' ')
	p.binary = data[binStart : binStart+binEnd]

	// Parse out ppid
	ppidPath := fmt.Sprintf("/proc/%d/ppid", p.pid)
	ppidBytes, err := ioutil.ReadFile(ppidPath)
	if err != nil {
		return err
	}
	ppidStr := strings.TrimSpace(string(ppidBytes))

	ppid, err := strconv.ParseInt(ppidStr, 10, 0)
	if err != nil {
		return err
	}
	p.ppid = int(ppid)

	return err
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

	return newPlan9Process(pid)
}

func processes() ([]Process, error) {
	d, err := os.Open("/proc")
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

			p, err := newPlan9Process(int(pid))
			if err != nil {
				continue
			}

			results = append(results, p)
		}
	}

	return results, nil
}

func newPlan9Process(pid int) (*Plan9Process, error) {
	p := &Plan9Process{pid: pid}
	return p, p.Refresh()
}
