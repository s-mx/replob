package containers

import "math/rand"
// TODO: проверить весь код здесь, после изменения int64 на int
type Set int

func NewSet(numberNodes uint32) Set {
	// Initialize first N-1 smallest bits by 1
	return (1<<numberNodes) - 1
}

func NewSetFromValue(value uint32) Set {
	return 1 << value
}

func NewRandomSubset(set Set, sizeSubset int, generator *rand.Rand) Set {
	subset := Set(0)
	sizeSet := int(set.Size())
	for sizeSubset > 0 {
		sizeSubset--
		for {
			ind := generator.Intn(sizeSet)
			if subset.Consist(uint32(ind)) == false {
				subset.Insert(uint32(ind))
				break
			}
		}
	}

	return subset
}

func (set Set) Size() uint32 {
	result := uint32(0)
	one := uint64(1)
	for ind := 0; ind < 32; ind++ {
		if (uint64(set) & (one << uint(ind))) > 0 {
			result++
		}
	}

	return result
}

func (set Set) Get(ind uint32) uint32 {
	var counter, indexSet uint32
	one := uint64(1)
	for counter < ind {
		if (uint64(set) & (one << uint(indexSet))) > 0 {
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
	*set = 0
}

func (set Set) Consist(elem uint32) bool {
	if elem >= 64 {
		return false
	}

	one := uint64(1)
	return (uint64(set) & (one << elem)) > 0
}

func (set Set) Equal(rgh Set) bool {
	return set == rgh
}

func (set Set) NotEqual(rgh Set) bool {
	return !set.Equal(rgh)
}

func (set *Set) AddSet(rghSet Set) {
	*set |= rghSet
}

func (set *Set) Insert(id uint32) {
	*set |= Set(uint64(1) << id)
}

func (set *Set) Intersect(rgh Set) {
	*set &= rgh
}

func (set *Set) Erase(arg interface{}) {
	id := arg.(uint32)
	if id < 32 && *set & (1 << id) > 0 {
		*set ^= 1 << id
	}
}

func (set Set) Diff(rgh Set) Set {
	return (set | rgh) ^ rgh
}
