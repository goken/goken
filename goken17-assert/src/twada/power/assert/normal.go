package assert

import (
	"testing"
)

func Ok(t *testing.T, actual bool) {
	if !actual {
		t.Errorf("failed")
	}
}
