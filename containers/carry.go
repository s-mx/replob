package containers

import (
	"sort"
	"encoding/binary"
)

// Types of carries
const (
	ALGORITHM_CARRY   = "ALGORITHM" // кажется, что не нужно
	MEMBERSHIP_CHANGE = "MEMBERSHIP_CHANGE"
)

type Payload interface {
	Type() string
	Bytes() []byte
}

type SimpleInt struct {
	value int
}

func NewSimpleInt(value int) *SimpleInt {
	return &SimpleInt{value:value,}
}

func (obj *SimpleInt) Type() string {
	return "Int32"
}

func (obj *SimpleInt) Bytes() []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(obj.value))
	return bs
}

type MembershipChangeCarry struct {
	//TODO: Пока что не ясно что нужно
}

func NewMembershipChangeCarry() *MembershipChangeCarry {
	return &MembershipChangeCarry{}
}

func (obj *MembershipChangeCarry) Type() string {
	return MEMBERSHIP_CHANGE
}

func (obj *MembershipChangeCarry) Bytes() []byte {
	return make([]byte, 1)
}

type ElementaryCarry struct {
	id		int
	value	Payload
}

func NewElementaryCarry(id int, val Payload) ElementaryCarry {
	return ElementaryCarry{
		id:id,
		value:val,
	}
}

func (obj *ElementaryCarry) GetId() int {
	return obj.id
}

func (obj *ElementaryCarry) GetPayload() Payload {
	return obj.value
}

type Carry struct {
	id    int
	value []ElementaryCarry
}

func (carry *Carry) GetId() int {
	return carry.id
}

func (carry *Carry) GetElementaryCarries() []ElementaryCarry {
	return carry.value
}

func (carry *Carry) Size() int {
	return len(carry.value)
}

func (carry *Carry) Append(elem ElementaryCarry) {
	carry.value = append(carry.value, elem)
}

func NewCarries(args ...ElementaryCarry) []Carry {
	result := make([]Carry, len(args))
	for ind := 0; ind < len(args); ind++ {
		result[ind] = NewCarry(ind, []ElementaryCarry{args[ind]})
	}

	return result
}

func NewCarriesN(number int) []Carry {
	result := make([]Carry, number)
	for ind := 0; ind < number; ind++ {
		elemCarry := NewElementaryCarry(ind, Payload(NewSimpleInt(number+1)))
		result[ind] = NewCarry(ind, []ElementaryCarry{elemCarry})
	}

	return result
}

func NewCarry(id int, val []ElementaryCarry) Carry {
	return Carry{
		id:    id,
		value: val,
	}
}

func (carry *Carry) Equal(otherCarry Carry) bool {
	return carry.id == otherCarry.id
}

func (carry *Carry) NotEqual(otherCarry Carry) bool {
	return !carry.Equal(otherCarry)
}

type CarriesSet struct {
	StepId   StepId
	ArrCarry []Carry
}

//type CarriesSet []Carry

func NewCarriesSet(args ...Carry) CarriesSet {
	ptr := new(CarriesSet)
	for _, val := range args {
		ptr.ArrCarry = append(ptr.ArrCarry, val)
	}

	sort.Sort(ById(ptr.ArrCarry))
	return *ptr
}

func (set *CarriesSet) SplitByType() (CarriesSet, CarriesSet) {
	algorithmCarries := NewCarriesSet()
	membershipCarries := NewCarriesSet()

	for _, carry := range set.ArrCarry {
		elemCarries := carry.GetElementaryCarries()
		algorithmCarry := NewCarry(carry.GetId(), make([]ElementaryCarry, 0))
		membershipCarry := NewCarry(carry.GetId(), make([]ElementaryCarry, 0))
		for _, elemCarry := range elemCarries {
			payload := elemCarry.GetPayload()
			if payload.Type() == MEMBERSHIP_CHANGE {
				membershipCarries.Append(carry)
			} else {
				algorithmCarry.Append(elemCarry) // TODO: придумать что-то получше
			}
		}

		if algorithmCarry.Size() > 0 {
			algorithmCarries.Append(algorithmCarry)
		}
		if membershipCarry.Size() > 0 {
			membershipCarries.Append(membershipCarry)
		}
	}

	return algorithmCarries, membershipCarries
}

func (set *CarriesSet) Equal(otherSet CarriesSet) bool {
	if set.Size() != otherSet.Size() {
		return false
	}

	for ind := 0; ind < set.Size(); ind++ {
		if set.ArrCarry[ind].NotEqual(otherSet.ArrCarry[ind]) {
			return false
		}
	}

	return true
}

func (set CarriesSet) NotEqual(otherSet CarriesSet) bool {
	return !set.Equal(otherSet)
}

func (set *CarriesSet) AddSet(otherSet CarriesSet) {
	for _, val := range otherSet.ArrCarry {
		set.Append(val)
	}
}

type ById []Carry

func (seq ById) Len() int {
	return len(seq)
}

func (seq ById) Less(i, j int) bool {
	return seq[i].id < seq[j].id
}

func (seq ById) Swap(i, j int) {
	seq[i], seq[j] = seq[j], seq[i]
}

func (set *CarriesSet) Consist(carry Carry) bool {
	for ind := 0; ind < set.Size(); ind++ {
		if set.ArrCarry[ind].Equal(carry) {
			return true
		}
	}

	return false
}

func (set *CarriesSet) Append(carry Carry) {
	if set.Consist(carry) {
		return
	}

	set.ArrCarry = append(set.ArrCarry, carry)
	sort.Sort(ById(set.ArrCarry))
}

func (set CarriesSet) Size() int {
	return len(set.ArrCarry)
}

func (set CarriesSet) Get(ind int) Carry {
	return set.ArrCarry[ind]
}

func (set *CarriesSet) Clear() {
	set.ArrCarry = make([]Carry, 0)
}
