package user

import (
	"runtime"
	"testing"
)

func TestCaptcha(t *testing.T) {
	_, fn, _, _ := runtime.Caller(0)
	t.Logf("%s\n", fn)
}
