package logger

import (
	"testing"
	"time"
)

func TestStd(t *testing.T) {
	Debug("1112345")
	Debugf("%s", "111")

	SetLoggerLevel(INFO)
	Info("3344")
	time.Sleep(3 * time.Second)
}
