package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
)

func TestOneNode(t *testing.T) {
	conf := NewMasterlessConfiguration(1)
	carry := cont.NewCarry(1)
	bc := NewSimpleBroadcaster(conf.Info)
	cm := NewSimpleCommitter(conf.Info)
	cons := NewConsensuser(bc, cm, conf, 0)

	cons.Propose(*carry)
	if cm.CheckLastCarry(carry) == false {
		t.Error("Carry isn't committed")
		t.Fail()
	}
}
