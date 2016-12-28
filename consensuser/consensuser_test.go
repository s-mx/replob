package consensuser

import (
	cont "github.com/s-mx/replob/containers"
	"testing"
)

func Run(t *testing.T) {
}

func TestSimple(t *testing.T) {
	conf := NewMasterlessConfiguration(1)
	carry := NewCarrier(0, 1)
	broadcaster := NewSimpleBroadcaster()
	committer := NewSimpleCommitter()
	cons := NewConsensuser(broadcaster, committer, conf, 0)

	cons.Propose(carry)
	for broadcaster.Length() > 0 {
		switch msg := broadcaster.Get(0); msg.typeMessage {
		case cont.Vote:
		case cont.Commit: // Кажется, будто не нужно
		case cont.Disconnect:
		}
	}
}
