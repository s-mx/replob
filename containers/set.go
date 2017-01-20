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
	return &Set{maskNodes: 1 << value}
}

func (set *Set) Size() uint32 {
	result := uint32(0)
	one := uint64(1)
	for ind := 0; ind < 64; ind++ {
		if (set.maskNodes & (one << uint(ind))) > 0 {
			result++
		}
	}

	return result
}

func (set *Set) Get(ind uint32) uint32 {
	var counter, indexSet uint32
	one := uint64(1)
	for counter < ind {
		if (set.maskNodes & (one << uint(indexSet))) > 0 {
			counter++
		}

		if counter == ind {
			return indexSet
		}

		indexSet++
	}

	return 0
}

func (set *Set) Clear() {
	set.maskNodes = 0
}

func (set *Set) Consist(elem uint32) bool {
	one := uint64(1)
	return (set.maskNodes & (one << elem)) > 0
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
