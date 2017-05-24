package lock

import (
	"github.com/s-mx/replob/containers"
	"encoding/gob"
	"bytes"
	"time"
)

type Message struct {
	TypeMessage string
	LockId      string
	ClientId    string
	Duration    time.Duration
}

type MessageCarry struct {
	TypeMessage string
	ArrBytes    []byte
}

func (message MessageCarry) Type() string {
	return message.TypeMessage
}

func (message MessageCarry) Bytes() []byte {
	return message.ArrBytes
}

func (message *Message) Marshall() bytes.Buffer {
	var buffer bytes.Buffer
	gob.NewEncoder(&buffer).Encode(message)
	return buffer
}

func Unmarshall(carry containers.ElementaryCarry) *Message {
	var message Message
	buffer := carry.GetPayload().Bytes()
	gob.NewDecoder(bytes.NewBuffer(buffer)).Decode(&message)
	return &message
}