// +build darwin

package ps

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindProcessDarwin(t *testing.T) {
	proc := testFindProcess(t, "go-ps.test")
	assert.True(t, proc.PPid() > 0)
}

func TestProcessesDarwin(t *testing.T) {
	testProcesses(t, "go")
}

func TestProcessesDarwinError(t *testing.T) {
	errFn := func() ([]Process, error) {
		return nil, fmt.Errorf("oops")
	}
	proc, err := findProcessWithFn(errFn, os.Getpid())
	assert.Nil(t, proc)
	assert.EqualError(t, err, "Error listing processes: oops")
}

func TestProcessExecRemoved(t *testing.T) {
	procPath, cmd, proc := testExecRun(t)
	defer cleanup(cmd, procPath)
	t.Logf("Ran with PID: %d", cmd.Process.Pid)
	// Remove it while it is running
	_ = os.Remove(procPath)
	matchPath := func(p Process) bool { return p.Pid() == proc.Pid() }
	procs, err := findProcessesWithFn(processes, matchPath, 1)
	require.NoError(t, err)
	t.Logf("Proc: %#v", procs[0])
}
