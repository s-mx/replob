package lock

import (
	"testing"
	"github.com/s-mx/replob/network"
	"time"
	"encoding/gob"
	"log"
)

type Context struct {
	conf		*network.Configuration
	dispatchers	[]*network.NetworkDispatcher
	replobs		[]*lockReplob
}

func CreateNodes(numberNodes int, t *testing.T) *Context {
	context := &Context{
		conf:network.NewLocalNetConfiguration(numberNodes),
		dispatchers:make([]*network.NetworkDispatcher, numberNodes),
		replobs:make([]*lockReplob, numberNodes),
	}

	for ind := 0; ind < numberNodes; ind++ {
		context.replobs[ind] = newLockReplob()
		curReplob := network.Replob(context.replobs[ind])
		curDisp, _ := network.NewConsensuser(ind, context.conf, curReplob)
		context.dispatchers[ind] = curDisp
		context.replobs[ind].dispatcher = context.dispatchers[ind]
	}

	return context
}

func TestOneLock(t *testing.T) {
	context := CreateNodes(2, t)
	context.dispatchers[0].Start()
	context.dispatchers[1].Start()
	context.dispatchers[2].Start()

	gob.Register(MessageCarry{})

	client1 := context.replobs[0].NewLock("client1")

	res, err := client1.AcquireLock("lock 1")
	if res != OK {
		log.Panicf("WRONG AcquireLock: %s", err.Error())
	}

	time.Sleep(500 * time.Microsecond)
	res, err = client1.Unlock("lock 1")
	if res != OK {
		log.Panicf("WRONG Unlock: %s", err.Error())
	}

	time.Sleep(5 * time.Second)

	context.dispatchers[0].Stop()
}

func TestUpdateLease(t *testing.T) {
	_ = CreateNodes(3, t)
}