package network

import (
	"net"
	"encoding/gob"
	cont "github.com/s-mx/replob/containers"
	"sync"
	"log"
	"time"
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

func (service *ClientService) start() {
	defer func() {
		service.connection.Close()
		service.waitGroup.Done()
	}()

	var raddr *net.TCPAddr
	var err error
	raddr, err = net.ResolveTCPAddr("tcp", service.service)
	service.connection, err = net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	for {
		select {
		case <-service.channelStop:
			return
		default:
		}

		var message cont.Message
		select {
		case message = <-service.channelMessage:
		case <-time.After(time.Second):
			continue
		}

		// FIXME: implement reconnection and appropriate error handling
		checkError(err)
		err = gob.NewEncoder(service.connection).Encode(message)
		// TODO: понять какие ошибки обрабатывать
		// Логируем все ошибки
		// проверяем io.EOF, Timeout
		// Ставим Timeout на запись, кастуем к OpError, проверяем Timeout()
		// пересоздаем соединение
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
