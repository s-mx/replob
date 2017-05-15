package lock

import (
	"github.com/s-mx/replob/containers"
	"encoding/gob"
	"bytes"
)

type Message struct {
	typeMessage		string
	lockId			string
	clientId		string
}

func Marshall(message *Message) bytes.Buffer {

}

func Unmarshall(carry containers.ElementaryCarry) *Message {

}