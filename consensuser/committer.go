package consensuser

type Committer interface {
	Commit(Carrier)
}

type MyCommitter struct {
}

func (commiter *MyCommitter) Commit(carrier Carrier) {

}

type SimpleCommitter struct {
}

func (committer *SimpleCommitter) Commit(carry Carrier) {
	committer.resultStatus[carry.id] = true
}

func (committer *SimpleCommitter) CommitSet(set CarriesSet) {
	for ind := 0; ind < set.Length(); ind++ {
		committer.Commit(set.Get(ind))
	}
}
