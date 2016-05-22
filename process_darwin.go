// +build darwin

package ps

/*
#include <stdio.h>
#include <errno.h>
#include <libproc.h>
extern void darwinProcesses();
*/
import "C"

import (
	"path/filepath"
	"sync"
)

// This lock is what verifies that C calling back into Go is only
// modifying data once at a time.
var darwinLock sync.Mutex
var darwinProcs []Process

// DarwinProcess is process definition for OS X
type DarwinProcess struct {
	pid  int
	ppid int
	path string
}

// Pid returns process id
func (p *DarwinProcess) Pid() int {
	return p.pid
}

// PPid returns parent process id
func (p *DarwinProcess) PPid() int {
	return p.ppid
}

// Executable returns process executable name
func (p *DarwinProcess) Executable() string {
	return filepath.Base(p.path)
}

// Path returns path to process executable
func (p *DarwinProcess) Path() (string, error) {
	return p.path, nil
}

//export goDarwinAppendProc
func goDarwinAppendProc(pid C.pid_t, ppid C.pid_t, comm *C.char) {
	proc := &DarwinProcess{
		pid:  int(pid),
		ppid: int(ppid),
		path: C.GoString(comm),
	}
	darwinProcs = append(darwinProcs, proc)
}

func findProcess(pid int) (Process, error) {
	ps, err := processes()
	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		if p.Pid() == pid {
			return p, nil
		}
	}

	return nil, nil
}

func processes() ([]Process, error) {
	darwinLock.Lock()
	defer darwinLock.Unlock()
	darwinProcs = make([]Process, 0, 50)

	// To ignore deadcode warning for goDarwinAppendProc
	_ = goDarwinAppendProc

	C.darwinProcesses()

	return darwinProcs, nil
}
