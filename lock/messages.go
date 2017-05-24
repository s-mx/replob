package lock

import (
	"github.com/s-mx/replob/containers"
	"encoding/gob"
	"bytes"
	"time"
)

type Message struct {
	typeMessage string
	lockId      string
	clientId    string
	duration    time.Duration
}

type MessageCarry struct {
	typeMessage string
	arrBytes	[]byte
}

func (message MessageCarry) Type() string {
	return message.typeMessage
}

func (message MessageCarry) Bytes() []byte {
	return message.arrBytes
}

func (message *Message) Marshall() bytes.Buffer {
	var buffer bytes.Buffer
	gob.NewEncoder(&buffer).Encode(message)
	return buffer
}

func Unmarshall(carry containers.ElementaryCarry) *Message {
	var message Message
	gob.NewDecoder(bytes.NewBuffer(carry.GetPayload().Bytes())).Decode(message)
	return &message
}