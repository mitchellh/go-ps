// +build linux netbsd solaris

package ps

import (
	"testing"
)

func TestUnixProcess_impl(t *testing.T) {
	var _ Process = new(UnixProcess)
}
