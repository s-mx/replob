package containers

type Set struct {
	// Just info about state of replica
	maskNodes uint64
}

func NewSet(numberNodes uint32) *Set {
	ptrSet := new(Set)
	for i := uint32(0); i < uint32(numberNodes); i++ {
		ptrSet.maskNodes = ptrSet.maskNodes | (1 << i)
	}

	return ptrSet
}

func NewSetFromValue(value uint32) *Set {
	ptrSet := new(Set)
	ptrSet.maskNodes = 1 << value
	return ptrSet
}

func (set *Set) Equal(rgh *Set) bool {
	return set.maskNodes == rgh.maskNodes
}

func (set *Set) NotEqual(rgh *Set) bool {
	return !set.Equal(rgh)
}

func (set *Set) AddSet(rghSet *Set) {
	set.maskNodes |= rghSet.maskNodes
}

func (set *Set) Insert(id uint32) {
	set.maskNodes |= uint64(1) << id
}

func (set *Set) Intersect(rgh *Set) {
	set.maskNodes &= rgh.maskNodes
}

func (set *Set) Erase(id uint32) {
	set.maskNodes ^= 1 << id
}

func (set *Set) Diff(rgh *Set) *Set {
	ptrSet := NewSet(0)
	ptrSet.maskNodes = (set.maskNodes | rgh.maskNodes) ^ rgh.maskNodes
	return ptrSet
}
