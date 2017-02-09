package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
)

func TestOneNode(t *testing.T) {
	conf := NewMasterlessConfiguration(1)
	carry := cont.NewCarry(1)
	LocalDispatcers := NewLocalDispatchers(1, conf)
	dsp := Dispatcher(LocalDispatcers[0])

	helper := newTestCommitHelper(1)
	cm := NewTestLocalCommitter(0, helper)
	cons := NewCalmConsensuser(&dsp, Committer(cm), conf, 0)

	cons.Propose(*carry)
	if cm.CheckLastCarry(0, *carry) == false {
		t.Error("Carry isn't committed")
	}
}

func TestTwoNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(2)
	carry := cont.NewCarry(1)
	LocalDispatchers := NewLocalDispatchers(2, conf)
	dsp1 := Dispatcher(LocalDispatchers[0])
	dsp2 := Dispatcher(LocalDispatchers[1])

	helper := newTestCommitHelper(2)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	cons1 := NewCalmConsensuser(&dsp1, Committer(cm1), conf, 0)
	LocalDispatchers[0].cons = cons1
	cons2 := NewCalmConsensuser(&dsp2, Committer(cm2), conf, 1)
	LocalDispatchers[1].cons = cons2

	var err error
	cons1.Propose(*carry)
	err = LocalDispatchers[0].proceedFirstMessage(1)
	if err != nil {
		t.Error("Broadcast isn't correct")
	}

	err = LocalDispatchers[1].proceedFirstMessage(0)
	if err != nil {
		t.Error("Queue of messages doesn't contains the element")
	}

	if cm1.CheckLastCarry(0, *carry) == false ||
	   cm2.CheckLastCarry(1, *carry) == false {
		t.Error("Safety is broken")
	}
}


func TestThreeNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(3)
	carry := cont.NewCarry(1)
	LocalBroadcasters := NewLocalDispatchers(3, conf)
	dsp1 := Dispatcher(LocalBroadcasters[0])
	dsp2 := Dispatcher(LocalBroadcasters[1])
	dsp3 := Dispatcher(LocalBroadcasters[2])

	helper := newTestCommitHelper(3)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	cm3 := NewTestLocalCommitter(2, helper)

	cons1 := NewCalmConsensuser(&dsp1, Committer(cm1), conf, 0)
	cons2 := NewCalmConsensuser(&dsp2, Committer(cm2), conf, 1)
	cons3 := NewCalmConsensuser(&dsp3, Committer(cm3), conf, 2)
	LocalBroadcasters[0].cons = cons1
	LocalBroadcasters[1].cons = cons2
	LocalBroadcasters[2].cons = cons3

	var err error
	cons1.Propose(*carry)
	err = LocalBroadcasters[0].proceedFirstMessage(1)
	if err != nil {

	}

	err = LocalBroadcasters[0].proceedFirstMessage(2)
	if err != nil {

	}

	err = LocalBroadcasters[1].proceedFirstMessage(2)
	if err != nil {

	}

	err = LocalBroadcasters[2].proceedFirstMessage(1)
	if err != nil {

	}
}

/*
Tests TODO:
1. Check commit messages for 2 nodes.
2. 3 nodes.
3. Random tests.
 */
