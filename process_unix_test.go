// +build linux solaris

package ps

import (
	"os"
	"strings"
	"testing"
)

func TestUnixProcess_impl(t *testing.T) {
	var _ Process = new(UnixProcess)
}

func TestUnixProcessCmdLine(t *testing.T) {
	p, err := FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if p == nil {
		t.Fatal("should have process")
	}

	if p.Pid() != os.Getpid() {
		t.Fatalf("bad: %#v", p.Pid())
	}

	up, ok := p.(*UnixProcess)
	if !ok {
		t.Fatal("type assertion should be ok")
	}

	if !strings.Contains(up.CMDLine(), "go") {
		t.Fatal("should be go process")
	}
}
