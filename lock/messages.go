package lock

import (
	"github.com/s-mx/replob/containers"
	"encoding/gob"
	"bytes"
	"time"
)

type Message struct {
	typeMessage		string
	lockId			string
	clientId		string
	timeStamp		time.Time
}

func Marshall(message *Message) bytes.Buffer {

}

func Unmarshall(carry containers.ElementaryCarry) *Message {

}