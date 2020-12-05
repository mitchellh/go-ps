// +build darwin

package ps

// #include <libproc.h>
// #include <stdlib.h>
// #include <errno.h>
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

type DarwinProcess struct {
	pid    int
	ppid   int
	binary string
}

func (p *DarwinProcess) Pid() int {
	return p.pid
}

func (p *DarwinProcess) PPid() int {
	return p.ppid
}

func (p *DarwinProcess) Executable() string {
	return p.binary
}

// bufSize references the constant that the implementation
// of proc_pidpath uses under the hood to make sure that
// no overflows happen.
//
// See https://opensource.apple.com/source/xnu/xnu-2782.40.9/libsyscall/wrappers/libproc/libproc.c
const bufSize = C.PROC_PIDPATHINFO_MAXSIZE

func getExePathFromPid(pid int) (path string, err error) {
	// Allocate in the C heap a string (char* terminated
	// with `/0`) of size `bufSize` and then make sure
	// that we free that memory that gets allocated
	// in C (see the `defer` below).
	buf := C.CString(string(make([]byte, bufSize)))
	defer C.free(unsafe.Pointer(buf))

	// Call the C function `proc_pidpath` from the included
	// header file (libproc.h).
	ret, err := C.proc_pidpath(C.int(pid), unsafe.Pointer(buf), bufSize)
	if ret <= 0 {
		err = fmt.Errorf("failed to retrieve pid path: %v", err)
		return
	}

	// Convert the C string back to a Go string.
	path = C.GoString(buf)
	return
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
	buf, err := darwinSyscall()
	if err != nil {
		return nil, err
	}

	procs := make([]*kinfoProc, 0, 50)
	k := 0
	for i := _KINFO_STRUCT_SIZE; i < buf.Len(); i += _KINFO_STRUCT_SIZE {
		proc := &kinfoProc{}
		err = binary.Read(bytes.NewBuffer(buf.Bytes()[k:i]), binary.LittleEndian, proc)
		if err != nil {
			return nil, err
		}

		k = i
		procs = append(procs, proc)
	}

	darwinProcs := make([]Process, len(procs))
	for i, p := range procs {
		comm, err := getExePathFromPid(int(p.Pid))

		if err != nil {
			// Falls back to the kinfo_proc->kp_proc.p_comm if no string for the pid was found
			comm = darwinCstring(p.Comm)
		} else {
			// returns the last element of the process execution path
			comm = filepath.Base(comm)
		}

		darwinProcs[i] = &DarwinProcess{
			pid:    int(p.Pid),
			ppid:   int(p.PPid),
			binary: comm,
		}
	}

	return darwinProcs, nil
}

func darwinCstring(s [16]byte) string {
	i := 0
	for _, b := range s {
		if b != 0 {
			i++
		} else {
			break
		}
	}

	return string(s[:i])
}

func darwinSyscall() (*bytes.Buffer, error) {
	mib := [4]int32{_CTRL_KERN, _KERN_PROC, _KERN_PROC_ALL, 0}
	size := uintptr(0)

	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
		0)

	if errno != 0 {
		return nil, errno
	}

	bs := make([]byte, size)
	_, _, errno = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		uintptr(unsafe.Pointer(&bs[0])),
		uintptr(unsafe.Pointer(&size)),
		0,
		0)

	if errno != 0 {
		return nil, errno
	}

	return bytes.NewBuffer(bs[0:size]), nil
}

const (
	_CTRL_KERN         = 1
	_KERN_PROC         = 14
	_KERN_PROC_ALL     = 0
	_KINFO_STRUCT_SIZE = 648
)

type kinfoProc struct {
	_    [40]byte
	Pid  int32
	_    [199]byte
	Comm [16]byte
	_    [301]byte
	PPid int32
	_    [84]byte
}
