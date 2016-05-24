package ps

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFindProcess(t *testing.T, name string) Process {
	proc, err := FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NotNil(t, proc)
	assert.Equal(t, os.Getpid(), proc.Pid())

	if name != "" {
		assert.Equal(t, name, proc.Executable())
		path, err := proc.Path()
		require.NoError(t, err)
		t.Logf("Path: %s", path)
		assert.True(t, strings.HasSuffix(path, string(os.PathSeparator)+name))
	}
	return proc
}

func testProcesses(t *testing.T, name string) {
	// This test works because there will always be SOME processes running
	procs, err := Processes()
	require.NoError(t, err)
	require.True(t, len(procs) > 0)

	if name != "" {
		found := false
		for _, p := range procs {
			if p.Executable() == name {
				found = true
				break
			}
		}
		assert.True(t, found)
	}
}

func TestFindProcess(t *testing.T) {
	testFindProcess(t, "")
}

func TestProcesses(t *testing.T) {
	testProcesses(t, "")
}
