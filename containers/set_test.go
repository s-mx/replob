package containers

import "testing"

// Just warm-up
func TestConstructorOneValue(t *testing.T) {
	set := NewSetFromValue(1)
	one := uint64(1)
	if set != Set(one<<1) {
		t.Fail()
	}

	set = NewSetFromValue(2)

	if set != Set(one<<2) {
		t.Fail()
	}

	set = NewSetFromValue(63)

	if set != Set(one<<63) {
		t.Fail()
	}
}

func TestInsertErase(t *testing.T) {
	set1 := NewSet(0)

	primes := []uint32{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	mask := uint64(0)
	for _, elem := range primes {
		set1.Insert(elem)
		mask |= uint64(1) << elem
		if set1 != Set(mask) {
			t.Fail()
		}
	}

	for ind := len(primes) - 1; ind >= 0; ind-- {
		set1.Erase(primes[ind])
		mask ^= uint64(1) << primes[ind]
		if set1 != Set(mask) {
			t.Fail()
		}
	}
}
