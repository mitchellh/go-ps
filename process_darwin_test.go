// +build darwin

package ps

import "testing"

func TestFindProcessDarwin(t *testing.T) {
	testFindProcess(t, "go-ps.test")
}

func TestProcessesWindows(t *testing.T) {
	testProcesses(t, "go")
}
