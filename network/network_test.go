package network

import (
	"testing"
	"time"
	cont "github.com/s-mx/replob/containers"
)

func TestSendSimpleMessage(t *testing.T) {
	message := cont.NewEmptyMessage()

	NewLocalServer(":2000")
	SendMessage(":2000", message)
	time.Sleep(time.Second)
}
