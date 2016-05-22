// +build windows

package ps

import "testing"

func TestFindProcessWindows(t *testing.T) {
	testFindProcess(t, "go-ps.test.exe")
}

func TestProcessesWindows(t *testing.T) {
	testProcesses(t, "go.exe")
}
