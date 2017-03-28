package network

import (
	"net"
	"encoding/gob"
	cont "github.com/s-mx/replob/containers"
	"sync"
	"log"
	"time"
	"bytes"
)

const (
	OK = iota
	ERROR
)

type ClientService struct {
	id				int
	service			string
	channelMessage	chan cont.Message
	waitGroup		sync.WaitGroup

	connection		*net.TCPConn

	channelStop		chan interface{}

	isRunning		bool
}

func NewClientService(id int, service string) *ClientService {
	return &ClientService{
		id:id,
		service:service,
		channelStop:make(chan interface{}),
		channelMessage:make(chan cont.Message, 10), // FIXME: use flags instead
		isRunning:false,
	}
}

func (service *ClientService) loop() int {
	defer service.connection.Close()

	for {
		select {
		case <-service.channelStop:
			return OK
		default:
		}

		var message cont.Message
		select {
		case message = <-service.channelMessage:
		case <-time.After(time.Second):
			continue
		}

		var err error
		var buffer bytes.Buffer
		err = gob.NewEncoder(&buffer).Encode(message)
		checkError(err)

		service.connection.SetDeadline(time.Now().Add(time.Second))
		service.connection.Write(buffer.Bytes())
		if Err, ok := err.(*net.OpError); ok {
			if Err.Timeout() {
				log.Printf("Client [%d]: Timeout error", service.id)
				continue
			}

			log.Printf("Client [%d]: error %s", service.id, Err.Error())
			return ERROR
		}

		checkError(err)
		// TODO: понять какие ошибки обрабатывать
		// Логируем все ошибки
		// проверяем io.EOF, Timeout
		// Ставим Timeout на запись, кастуем к OpError, проверяем Timeout()
		// пересоздаем соединение
	}
}

func (service *ClientService) start() {
	defer func() {
		log.Printf("Client [%d]: stop working", service.id)
		service.waitGroup.Done()
	}()

	numberAttempts := 0
	for {
		numberAttempts++
		var raddr *net.TCPAddr
		var err1, err2 error
		raddr, err1 = net.ResolveTCPAddr("tcp", service.service)
		service.connection, err2 = net.DialTCP("tcp", nil, raddr)
		if err1 != nil || err2 != nil {
			if err1 != nil {
				log.Printf(err1.Error())
			} else {
				log.Printf(err2.Error())
			}

			if numberAttempts == int(1e9) {
				log.Printf("Client [%d]: Attempts to connect ended", service.id)
				return
			}

			log.Printf("Client [%d]: another %d attempt to connect", service.id, numberAttempts)
			continue
		}

		numberAttempts = 0
		if service.loop() == 0 {
			break
		}
	}
}

func (service *ClientService) Start() {
	service.isRunning = true
	log.Printf("Client [%d] just started", service.id)
	service.waitGroup.Add(1)
	go service.start()
}

func (service *ClientService) Stop() {
	if service.isRunning == false {
		return
	}

	service.isRunning = false
	service.channelStop<-0
	service.waitGroup.Wait()
	log.Printf("Client [%d]: stop working", service.id)
}
