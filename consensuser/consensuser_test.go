package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
)

func TestOneNode(t *testing.T) {
	conf := NewMasterlessConfiguration(1)
	carry := cont.NewCarry(1)
	bc := Broadcaster(NewSimpleBroadcaster(conf.Info))
	cm := NewSimpleCommitter(conf.Info)
	tmp := Committer(cm)
	cons := NewConsensuser(&bc, &tmp, conf, 0)

	cons.Propose(*carry)
	if cm.CheckLastCarry(0, carry) == false {
		t.Error("Carry isn't committed")
		t.Fail()
	}
}

func TestTwoNodes(t *testing.T) {
	conf := NewMasterlessConfiguration(2)
	carry := cont.NewCarry(1)
	bc := NewSimpleBroadcaster(conf.Info)
	cm := NewSimpleCommitter(conf.Info)
	tmpBc := Broadcaster(bc)
	tmpCm := Committer(cm)
	cons1 := NewConsensuser(&tmpBc, &tmpCm, conf, 0)
	cons2 := NewConsensuser(&tmpBc, &tmpCm, conf, 1)

	var err error
	cons1.Propose(*carry)
	err = bc.proceedMessage(cons2)
	if err != nil {
		t.Error("Broadcast isn't correct")
	}

	err = bc.proceedMessage(cons1)
	if err != nil {
		t.Error("abc:)")
	}

	if cm.CheckLastCarry(0, carry) == false ||
	   cm.CheckLastCarry(1, carry) == false {
		t.Error("Safety is broken")
	}
}
