// +build openbsd

package ps

// #cgo LDFLAGS: -lkvm
// #include "process_openbsd.h"
import "C"

var openbsdProcs []Process

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

type OpenBSDProcess struct {
	pid    int
	ppid   int
	binary string
}

func newOpenBSDProcess() *OpenBSDProcess {
	return &OpenBSDProcess{}
}

func (p *OpenBSDProcess) Pid() int {
	return p.pid
}

func (p *OpenBSDProcess) PPid() int {
	return p.ppid
}

func (p *OpenBSDProcess) Executable() string {
	return p.binary
}

//export go_openbsd_append_proc
func go_openbsd_append_proc(pid C.pid_t, ppid C.pid_t, comm *C.char) {
	proc := &OpenBSDProcess{
		pid:    int(pid),
		ppid:   int(ppid),
		binary: C.GoString(comm),
	}

	openbsdProcs = append(openbsdProcs, proc)
}

func processes() ([]Process, error) {
	openbsdProcs = make([]Process, 0, 50)

	_, err := C.openbsdProcesses()
	if err != nil {
		return nil, err
	}

	return openbsdProcs, nil
}
