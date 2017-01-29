package containers

import "sort"

var countId int
type Payload int

type Carry struct {
	Id    int
	value Payload // payload
}

func NewCarry(val int) *Carry {
	ptr := new(Carry)
	ptr.Id = countId
	countId++
	ptr.value = Payload(val)
	return ptr
}

func (carry *Carry) Equal(otherCarry Carry) bool {
	return carry.Id == otherCarry.Id
}

func (carry *Carry) NotEqual(otherCarry Carry) bool {
	return !carry.Equal(otherCarry)
}

type CarriesSet []Carry


func NewCarriesSet(args ...Carry) *CarriesSet {
	ptr := new(CarriesSet)
	for _, val := range args {
		*ptr = append(*ptr, val)
	}

	sort.Sort(ById(*ptr))
	return ptr
}

func (set CarriesSet) Equal(otherSet CarriesSet) bool {
	if len(set) != len(otherSet) {
		return false
	}

	for ind := 0; ind < len(set); ind++ {
		if set[ind].NotEqual(otherSet[ind]) {
			return false
		}
	}

	return true
}

func (set CarriesSet) NotEqual(otherSet CarriesSet) bool {
	return !set.Equal(otherSet)
}

func (set *CarriesSet) AddSet(otherSet CarriesSet) {
	for _, val := range otherSet {
		set.Append(val)
	}
}

type ById []Carry

func (seq ById) Len() int {
	return len(seq)
}

func (seq ById) Less(i, j int) bool {
	return seq[i].Id < seq[j].Id
}

func (seq ById) Swap(i, j int) {
	seq[i], seq[j] = seq[j], seq[i]
}

func (set *CarriesSet) Append(carry Carry) {
	*set = append(*set, carry)
	sort.Sort(ById(*set))
}

func (set CarriesSet) Size() int {
	return len(set)
}

func (set CarriesSet) Get(ind int) Carry {
	return set[ind]
}

func (set *CarriesSet) Clear() {
	*set = make([]Carry, 0)
}
