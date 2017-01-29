package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
)

func TestOneNode(t *testing.T) {
	conf := NewMasterlessConfiguration(1)
	carry := cont.NewCarry(1)
	bc := Broadcaster(NewTestLocalBroadcaster(conf.Info))

	helper := newTestCommitHelper(1)
	cm := NewTestLocalCommitter(0, helper)
	cons := NewCalmConsensuser(bc, Committer(cm), conf, 0)

	cons.Propose(*carry)
	if cm.CheckLastCarry(0, *carry) == false {
		t.Error("Carry isn't committed")
	}
}

func TestTwoNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(2)
	carry := cont.NewCarry(1)
	bc := NewTestLocalBroadcaster(conf.Info)

	helper := newTestCommitHelper(2)
	cm1 := NewTestLocalCommitter(0, helper)
	cm2 := NewTestLocalCommitter(1, helper)
	cons1 := NewCalmConsensuser(bc, Committer(cm1), conf, 0)
	cons2 := NewCalmConsensuser(bc, Committer(cm2), conf, 1)

	var err error
	cons1.Propose(*carry)
	err = bc.proceedMessage(cons2)
	if err != nil {
		t.Error("Broadcast isn't correct")
	}

	err = bc.proceedMessage(cons1)
	if err != nil {
		t.Error("Queue of messages doesn't contains the element")
	}

	if cm1.CheckLastCarry(0, *carry) == false ||
	   cm2.CheckLastCarry(1, *carry) == false {
		t.Error("Safety is broken")
	}
}
