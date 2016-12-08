package containers

type Set struct {
    // Just info about state of replica
    flagNodes map[uint32]bool
}

func NewSet(numberNodes uint32) *Set {
    ptrSet := new(Set)
    ptrSet.flagNodes = make(map[uint32]bool)
    for i := uint32(0); i < uint32(numberNodes); i++ {
        ptrSet.flagNodes[i] = true
    }

    return ptrSet
}

func NewSetFromValue(value uint32) *Set {
    ptrSet := new(Set)
    ptrSet.flagNodes = make(map[uint32]bool)
    ptrSet.flagNodes[value] = true
    return ptrSet
}

func (set *Set) Change(id uint32, val bool) {
    set.flagNodes[id] = val
}

func (set *Set) Equal(rgh *Set) bool {
    return true // dummy
}

func (set *Set) NotEqual(rgh *Set) bool {
    return ! set.Equal(rgh)
}

func (set* Set) AddSet(rghSet *Set) {
    // dummy
}

func (set* Set) Insert(id uint32) {

}

func (set *Set) Intersect(otherSet *Set) {
    // dummy
}