// +build linux

package ps

import "testing"

func TestUnixProcess(t *testing.T) {
	var _ Process = new(UnixProcess)
}
