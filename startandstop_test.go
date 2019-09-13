package gopher

import (
	"testing"
	"time"
)

func TestStartAndStop(t *testing.T) {
	go Start(nil)
	time.Sleep(time.Second * 2)
	if sdErr := ShutDown(); sdErr != nil {
		t.Errorf(sdErr.Error())
	}
}
