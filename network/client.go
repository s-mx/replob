package network

import (
	"net"
	"encoding/gob"
	cont "github.com/s-mx/replob/containers"
	"sync"
	"log"
)

type ClientService struct {
	id				int
	service			string
	channelMessage	chan cont.Message
	waitGroup		sync.WaitGroup

	connection		*net.TCPConn

	isRunning		bool
	mutexRunning	sync.Mutex
}

func NewClientService(id int, service string) *ClientService {
	return &ClientService{
		id:id,
		service:service,
		channelMessage:make(chan cont.Message, 10), // FIXME: use flags instead
		isRunning:false,
	}
}

func (service *ClientService) start() {
	defer service.waitGroup.Done()

	var raddr *net.TCPAddr
	var err error
	raddr, err = net.ResolveTCPAddr("tcp", service.service)
	service.connection, err = net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	for {
		var message cont.Message
		var more bool
		select {
		case message, more = <-service.channelMessage: // FIXME: channel Stop
		default:
			continue
		}

		if more == false {
			log.Printf("INFO client server[%d]: stopping working\n", service.id)
			return
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
	defer service.mutexRunning.Unlock()
	service.mutexRunning.Lock()
	service.waitGroup.Add(1)
	go service.start()
}

func (service *ClientService) Stop() {
	defer service.mutexRunning.Unlock() // FIXME: channelStop
	service.mutexRunning.Lock()

	if service.isRunning == false {
		return
	}

	close(service.channelMessage)
	service.waitGroup.Wait()
}

func NewClient(service string) net.Conn {
	conn, err := net.Dial("tcp", service)
	checkError(err)
	return conn
}
