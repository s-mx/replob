package network

import (
	"testing"
	"github.com/s-mx/replob/containers"
	"time"
)

func TestSendSimpleConnection(t *testing.T) {
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

func TestSimpleMessage(t *testing.T) {
	conf := NewLocalNetConfiguration(2)
	client1 := NewClientService(0, conf.serviceServer[0])
	server1 := NewServerService(0, conf)

	msg := containers.NewEmptyMessage()

	server1.Start()
	client1.Start()

	select {
	case client1.channelMessage<-msg:
	case <-time.After(time.Second):
		t.Fatal("Message didn't send")
	}

	var msgRec containers.Message
	select {
	case msgRec = <-server1.channelMessage:
	case <-time.After(time.Second):
		t.Fatal("Message didn't received")
	}

	if msg.NotEqual(msgRec) {
		t.Fatal("Received message isn't correct")
	}

	client1.Stop()
	server1.Stop()
}

func TestTwoNodes(t *testing.T) {
	conf := NewLocalNetConfiguration(2)
	client1 := NewClientService(0, conf.serviceServer[0])
	client2 := NewClientService(1, conf.serviceServer[1])
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

func TestTwoConsensusers(t *testing.T) {
	// FIXME: Иногда падает, когда не может сделать bind из-за занятого порта

	simpleInts := []containers.ElementaryCarry{
		containers.NewElementaryCarry(1, containers.SimpleInt(1)),
		containers.NewElementaryCarry(2, containers.SimpleInt(2)),
		containers.NewElementaryCarry(3, containers.SimpleInt(3)),
	}

	values := containers.NewCarries(simpleInts[0], simpleInts[1], simpleInts[2])

	config := NewLocalNetConfiguration(2)
	replob1 := NewLocalReplob()
	replob2 := NewLocalReplob()
	disp1, _ := NewConsensuser(0, config, replob1)
	disp2, _ := NewConsensuser(1, config, replob2)

	disp1.Start()
	disp2.Start()

	replob1.Propose(values[0])
	replob2.Propose(values[1])
	replob1.Propose(values[2])

	//time.Sleep(5 * time.Second)

	disp1.StopWait()
	disp2.StopWait()
}