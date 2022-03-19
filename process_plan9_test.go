// +build plan9

package ps

import (
	"testing"
)

func TestPlan9Process_impl(t *testing.T) {
	var _ Process = new(Plan9Process)
}
