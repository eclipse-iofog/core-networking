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
	inner:
		for {
			select {
			case <-p.ackChannel:
				timer.Stop()
				break inner
			case <-timer.C:
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
	for {
		select {
		case msg := <-p.inMessage:
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
	b := make([]byte, 0, MAX_READ_BUFFER_SIZE)
	isBroken := false
	for {
		select {
		case <-done:
			return
		case data := <-p.out:
			switch string(data) {
			case ACK:
				p.ackChannel <- 0
			case TXEND:
				p.in <- []byte(ACK)
				if !isBroken {
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
