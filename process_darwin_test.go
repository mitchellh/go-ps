// +build darwin

package ps

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindProcessDarwin(t *testing.T) {
	proc := testFindProcess(t, "go-ps.test")
	assert.Equal(t, 0, proc.PPid())
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
