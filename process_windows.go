// +build windows

package ps

func findProcess(pid int) (Process, error) {
	panic("not implemented for Windows")
}

func processes() ([]Process, error) {
	panic("not implemented for Windows")
}
