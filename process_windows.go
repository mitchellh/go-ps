// +build windows

package ps

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// Windows API functions
var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCloseHandle              = modKernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procGetProcessTimes          = modKernel32.NewProc("GetProcessTimes")
	procFileTimeToSystemTime     = modKernel32.NewProc("FileTimeToSystemTime")
	procOpenProcess              = modKernel32.NewProc("OpenProcess")
	procProcess32First           = modKernel32.NewProc("Process32FirstW")
	procProcess32Next            = modKernel32.NewProc("Process32NextW")
)

// Some constants from the Windows API
const (
	ERROR_NO_MORE_FILES = 0x12
	MAX_PATH            = 260
	PROCESS_ALL_ACCESS  = 0x1F0FFF
)

// PROCESSENTRY32 is the Windows API structure that contains a process's
// information.
type PROCESSENTRY32 struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	CntThreads        uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [MAX_PATH]uint16
}

type HANDLE uintptr

type FILETIME struct {
	LowDateTime      uint32
	HighDateTime     uint32
}
type SYSTEMTIME struct {
	year, month, dow, day, hour, min, sec, msec uint16
}

// WindowsProcess is an implementation of Process for Windows.
type WindowsProcess struct {
	pid   int
	ppid  int
	exe   string
	ctime time.Time
}

func (p *WindowsProcess) Pid() int {
	return p.pid
}

func (p *WindowsProcess) PPid() int {
	return p.ppid
}

func (p *WindowsProcess) Executable() string {
	return p.exe
}

func (p *WindowsProcess) CreationTime() time.Time {
	return p.ctime
}

func newWindowsProcess(e *PROCESSENTRY32, ctime time.Time) *WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	return &WindowsProcess{
		pid:   int(e.ProcessID),
		ppid:  int(e.ParentProcessID),
		exe:   syscall.UTF16ToString(e.ExeFile[:end]),
		ctime: ctime,
	}
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
	handle, _, _ := procCreateToolhelp32Snapshot.Call(
		0x00000002,
		0)
	if handle < 0 {
		return nil, syscall.GetLastError()
	}
	defer procCloseHandle.Call(handle)

	var (
		entry PROCESSENTRY32
		ctime, etime, ktime, utime FILETIME
		// real creation time
		rCtime = SYSTEMTIME{0,0,0,0,0,0,0,0}
	)
	entry.Size = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("Error retrieving process info.")
	}

	results := make([]Process, 0, 50)
	for {
		ret, _, _ = procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		// All done iterating over processes
		if ret == 0 {
			break
		}

		// Try to open process to capture more process information like ctime
		pHandle, _, _ := procOpenProcess.Call(PROCESS_ALL_ACCESS, uintptr(0), uintptr(entry.ProcessID))
		if pHandle != 0 {
			ret, _, _    = procGetProcessTimes.Call(uintptr(unsafe.Pointer(pHandle)),
					uintptr(unsafe.Pointer(&ctime)),
					uintptr(unsafe.Pointer(&etime)),
					uintptr(unsafe.Pointer(&ktime)),
					uintptr(unsafe.Pointer(&utime)))
			if ret != 0 {
				ret, _, _ = procFileTimeToSystemTime.Call(uintptr(unsafe.Pointer(&ctime)), uintptr(unsafe.Pointer(&rCtime)))
			}
		} else {
			rCtime = SYSTEMTIME{0,0,0,0,0,0,0,0}
		}
		ctime := time.Date(int(rCtime.year), time.Month(rCtime.month), int(rCtime.day),
			int(rCtime.hour), int(rCtime.min), int(rCtime.sec), 0, &time.Location{})

		results = append(results, newWindowsProcess(&entry, ctime))

		//fmt.Printf("process age over? %v\n", time.Since(pDate) > time.Duration(1 * time.Hour))

	}

	return results, nil
}
