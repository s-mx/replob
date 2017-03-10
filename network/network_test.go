package network

import (
	"testing"
)

func TestSendSimpleMessage(t *testing.T) {
	conf := NewLocalNetConfiguration(2)
	client1 := NewClientService(0, conf.serviceServer[1])
	client2 := NewClientService(1, conf.serviceServer[0])
	server1 := NewServerService(0, conf)
	server2 := NewServerService(1, conf)

	server1.Start()
	server2.Start()
	client1.Start()
	client2.Start()

	client1.Stop()
	client2.Stop()
	server1.Stop()
	server2.Stop()
}
