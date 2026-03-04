package stacktrace

import (
	"testing"

	"github.com/pkg/errors"
)

func TestGetStackStraceNil(t *testing.T) {
	st := GetStackStrace(nil)
	if len(st) != 0 {
		t.Fatalf("expected empty slice for nil error, got %v", st)
	}
}

func TestGetStackStraceWrappedError(t *testing.T) {
	base := errors.New("boom")
	wrapped := errors.Wrap(base, "wrapped")
	st := GetStackStrace(wrapped)
	if len(st) == 0 {
		t.Fatalf("expected stack frames for wrapped error, got none")
	}
	for i, s := range st {
		if s == "" {
			t.Fatalf("frame %d is empty", i)
		}
	}
}
