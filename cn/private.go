package cn

import (
	"crypto/tls"
	sdk "github.com/iotracks/container-sdk-go"
	"time"
)

type PrivateConnection struct {
	*ComSatConn

	ackChannel    chan byte
	newMsgChannel chan byte
	inMessage     chan *sdk.IoMessage
	outMessage    chan *sdk.IoMessage
	bytesToSend   []byte
	readyConn     chan<- Connector
}

func newPrivateConnection(id int,
	address, passcode string,
	hbInterval, hbThreshold time.Duration,
	tlsConfig *tls.Config,
	ready chan<- Connector) *PrivateConnection {
	return &PrivateConnection{
		ComSatConn:    newConn(id, address, passcode, hbInterval, hbThreshold, tlsConfig),
		ackChannel:    make(chan byte, 1),
		newMsgChannel: make(chan byte, 1),
		inMessage:     make(chan *sdk.IoMessage, READ_CHANNEL_BUFFER_SIZE),
		outMessage:    make(chan *sdk.IoMessage, WRITE_CHANNEL_BUFFER_SIZE),
		readyConn:     ready,
	}
}

func (p *PrivateConnection) Connect() {
	go p.ComSatConn.Connect()
	done := make(chan byte)
	go p.writeConnection(done)
	go p.readConnection(done)
	go p.monitorMessageDelivery(done)
	p.readyConn <- p
	select {
	case <-p.done:
		logger.Printf("[ PrivateConnection #%d ] Stopped by demand\n", p.id)
		close(done)
		return
	}
}

func (p *PrivateConnection) Disconnect() {
	p.ComSatConn.Disconnect()
	p.done <- 0
	p.bytesToSend = nil
}

func (p *PrivateConnection) monitorMessageDelivery(done <-chan byte) {
	logger.Printf("[ PrivateConnection #%d ] msg monitor goroutine started. isBusy = %t\n", p.id, p.bytesToSend != nil)
	defer logger.Printf("[ PrivateConnection #%d ] msg monitor goroutine exited. isBusy = %t\n", p.id, p.bytesToSend != nil)
	if p.bytesToSend != nil {
		p.pushMessageBytes()
		p.newMsgChannel <- 0
	}
	for {
		select {
		case <-done:
			return
		case <-p.newMsgChannel:
		}
		attempt := uint(0)
		timer := time.NewTimer(RETRY_SEND_TIMEOUT)
		logger.Printf("[ PrivateConnection #%d ] Has sent message\n", p.id)
	inner:
		for {
			select {
			case <-p.ackChannel:
				logger.Printf("[ PrivateConnection #%d ] Has sent message successfully\n", p.id)
				timer.Stop()
				break inner
			case <-timer.C:
				logger.Printf("[ PrivateConnection #%d ] Retrying to send message\n", p.id)
				if attempt < SEND_ATTEMPT_LIMIT {
					attempt++
				} else {
					logger.Printf("[ PrivateConnection #%d ] Failed to send message\n", p.id)
					break inner
				}
				p.pushMessageBytes()
				timer.Reset(1 << attempt * RETRY_SEND_TIMEOUT)
			case <-done:
				timer.Stop()
				return
			}
		}
		p.bytesToSend = nil
		p.readyConn <- p
	}
}

func (p *PrivateConnection) writeConnection(done <-chan byte) {
	logger.Printf("[ PrivateConnection #%d ] write goroutine started\n", p.id)
	defer logger.Printf("[ PrivateConnection #%d ] write goroutine exited\n", p.id)
	for {
		select {
		case msg := <-p.inMessage:
			logger.Printf("[ PrivateConnection #%d ] got message to write\n", p.id)
			if bytes, err := sdk.PrepareMessageForSendingViaSocket(msg); err != nil {
				logger.Printf("[ PrivateConnection #%d ] Error while encoding message: %s\n", p.id, err.Error())
			} else {
				p.bytesToSend = bytes
				p.pushMessageBytes()
				p.newMsgChannel <- 0
			}
		case <-done:
			return
		}
	}
}

func (p *PrivateConnection) readConnection(done <-chan byte) {
	logger.Printf("[ PrivateConnection #%d ] read goroutine started\n", p.id)
	defer logger.Printf("[ PrivateConnection #%d ] read goroutine exited\n", p.id)
	b := make([]byte, 0, MAX_READ_BUFFER_SIZE)
	isBroken := false
	for {
		select {
		case <-done:
			return
		case data := <-p.out:
			logger.Printf("[ PrivateConnection #%d ] Has read %d bytes %v\n", p.id, len(data), data)
			switch string(data) {
			case ACK:
				p.ackChannel <- 0
			case TXEND:
				logger.Printf("[ PrivateConnection #%d ] Sending ACK...\n", p.id)
				p.in <- []byte(ACK)
				logger.Printf("[ PrivateConnection #%d ] Has sent ACK\n", p.id)
				if !isBroken {
					logger.Printf("[ PrivateConnection #%d ] Going to parse %s\n", p.id, b)
					if msg, err := sdk.GetMessageReceivedViaSocket(b); err != nil {
						logger.Printf("[ PrivateConnection #%d ] Error while decoding message: %s", p.id, err.Error())
					} else {
						p.outMessage <- msg
					}
				}
				b = b[:0]
				isBroken = false
			default:
				if len(b)+len(data) <= MAX_READ_BUFFER_SIZE {
					b = append(b, data...)
				} else {
					isBroken = true
				}
			}
		}
	}
}

func (p *PrivateConnection) pushMessageBytes() {
	if p.isConnected {
		p.in <- p.bytesToSend
		p.in <- []byte(TXEND)
	}
}
