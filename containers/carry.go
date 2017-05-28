package containers

import (
	"encoding/binary"
	"errors"
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
	Id    int64 //TODO: move to int64
	Value Payload
}

func CreateElementaryId(nodeId int, carryId int) int64 {
	return int64(nodeId) << 32 + int64(carryId)
}

func NewElementaryCarry(id int64, val Payload) ElementaryCarry {
	return ElementaryCarry{
		Id:    id,
		Value: val,
	}
}

func (obj *ElementaryCarry) GetId() int64 {
	return obj.Id
}

func (obj *ElementaryCarry) Less(rgh *ElementaryCarry) bool {
	return obj.Id < rgh.Id // TODO: уточнить это
}

func (obj *ElementaryCarry) GetPayload() Payload {
	return obj.Value
}

func (obj *ElementaryCarry) Equal(rgh *ElementaryCarry) bool {
	return obj.Id == rgh.Id
}

func (obj *ElementaryCarry) NotEqual(rgh *ElementaryCarry) bool {
	return ! obj.Equal(rgh)
}

type Carry struct {
	ArrCarry []ElementaryCarry
}

func (carry *Carry) GetElementaryCarries() []ElementaryCarry {
	return carry.ArrCarry
}

func (carry *Carry) Size() int {
	return len(carry.ArrCarry)
}

func (carry *Carry) Clear() {
	carry.ArrCarry = make([]ElementaryCarry, 0)
}

func (carry *Carry) GetCarry(ind int) (*Carry, error) {
	if ind < 0 || ind >= carry.Size() {
		return nil, errors.New("out of range")
	}

	return NewCarry([]ElementaryCarry{carry.ArrCarry[ind]}), nil
}

func (carry *Carry) Get(ind int) (*ElementaryCarry, error) {
	if ind < 0 || ind >= carry.Size() {
		return nil, errors.New("out of range")
	}

	return &carry.ArrCarry[ind], nil
}

func (carry *Carry) GetFirst() (*ElementaryCarry, error) {
	return carry.Get(0)
}

func (carry *Carry) Append(elem *ElementaryCarry) {
	carry.ArrCarry = append(carry.ArrCarry, *elem)
}

func (carry *Carry) Merge(rghCarry *Carry) *Carry {
	res := NewCarry([]ElementaryCarry{})
	ind1, ind2 := 0, 0
	for ind1 < carry.Size() && ind2 < rghCarry.Size() {
		elem1, _ := carry.Get(ind1)
		elem2, _ := carry.Get(ind2)
		if elem1.Less(elem2) {
			res.Append(elem1)
			ind1++
		} else {
			res.Append(elem2)
			ind2++
		}
	}

	for ind1 < carry.Size() {
		elem, _ := carry.Get(ind1)
		res.Append(elem)
		ind1++
	}

	for ind2 < rghCarry.Size() {
		elem, _ := rghCarry.Get(ind2)
		res.Append(elem)
		ind2++
	}

	return res
}

func (carry *Carry) Union(rgh *Carry) *Carry {
	if carry.Size() + rgh.Size() == 0 {
		return NewCarry([]ElementaryCarry{})
	}
	merged := carry.Merge(rgh)
	elemCarries := merged.GetElementaryCarries()
	res := NewCarry([]ElementaryCarry{elemCarries[0]})
	for ind := 1; ind < merged.Size(); ind++ {
		if elemCarries[ind].NotEqual(&elemCarries[ind - 1]) {
			res.Append(&elemCarries[ind])
		}
	}

	return res
}

func NewCarries(args ...ElementaryCarry) []Carry {
	result := make([]Carry, len(args))
	for ind := 0; ind < len(args); ind++ {
		result[ind] = *NewCarry([]ElementaryCarry{args[ind]})
	}

	return result
}

// TODO: не удобно использовать при разных nodeId
func NewIntCarriesN(number int, nodeId int) Carry {
	result := Carry{
		ArrCarry: make([]ElementaryCarry, number),
	}
	for ind := 0; ind < number; ind++ {
		result.ArrCarry[ind] = NewElementaryCarry(CreateElementaryId(nodeId, ind), Payload(NewSimpleInt(number+1)))
	}

	return result
}

func NewCarry(val []ElementaryCarry) *Carry {
	return &Carry{
		ArrCarry: val,
	}
}

func (carry *Carry) Equal(otherCarry Carry) bool {
	if len(carry.ArrCarry) != len(otherCarry.ArrCarry) {
		return false
	}

	length := len(carry.ArrCarry)
	for ind := 0; ind < length; ind++ {
		if carry.ArrCarry[ind].Id != otherCarry.ArrCarry[ind].Id {
			return false
		}
	}

	return true
}

func (carry *Carry) NotEqual(otherCarry Carry) bool {
	return !carry.Equal(otherCarry)
}

func (carry *Carry) SplitByType() (alg *Carry, membership *Carry) {
	algorithmCarries := NewCarry([]ElementaryCarry{})
	membershipCarries := NewCarry([]ElementaryCarry{})

	for _, elemCarry := range carry.ArrCarry {
		payload := elemCarry.Value
		if payload.Type() == MEMBERSHIP_CHANGE {
			membershipCarries.Append(&elemCarry)
		} else {
			algorithmCarries.Append(&elemCarry)
		}
	}

	return algorithmCarries, membershipCarries
}
