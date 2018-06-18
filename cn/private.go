package cn

import (
	"crypto/tls"
	sdk "github.com/ioFog/iofog-go-sdk"
	"time"
)

type PrivateConnection struct {
	*ComSatConn

	inMessage  chan *sdk.IoMessage
	outMessage chan *sdk.IoMessage
	readyConn  chan<- Connector
}

func newPrivateConnection(id int,
	address, passcode string,
	hbInterval, hbThreshold time.Duration,
	tlsConfig *tls.Config,
	ready chan<- Connector) *PrivateConnection {
	return &PrivateConnection{
		ComSatConn: newConn(id, address, passcode, hbInterval, hbThreshold, tlsConfig),
		inMessage:  make(chan *sdk.IoMessage, READ_CHANNEL_BUFFER_SIZE),
		outMessage: make(chan *sdk.IoMessage, WRITE_CHANNEL_BUFFER_SIZE),
		readyConn:  ready,
	}
}

func (p *PrivateConnection) Connect() {
	go p.ComSatConn.Connect()
	done := make(chan byte)
	go p.writeConnection(done)
	go p.readConnection(done)
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
}

func (p *PrivateConnection) writeConnection(done <-chan byte) {
	for {
		select {
		case msg := <-p.inMessage:
			if bytes, err := sdk.PrepareMessageForSendingViaSocket(msg); err != nil {
				logger.Printf("[ PrivateConnection #%d ] Error while encoding message: %s\n", p.id, err.Error())
			} else {
				p.in <- bytes
				p.in <- []byte(TXEND)
			}
			p.readyConn <- p
		case <-done:
			return
		}
	}
}

func (p *PrivateConnection) readConnection(done <-chan byte) {
	b := make([]byte, 0, MAX_READ_BUFFER_SIZE)
	isBroken := false
	addToBuffer := func(bytes []byte) {
		if len(b)+len(bytes) <= MAX_READ_BUFFER_SIZE {
			b = append(b, bytes...)
		} else {
			isBroken = true
		}
	}
	for {
		select {
		case <-done:
			return
		case data := <-p.out:
			end := make([]byte, 5) // TXEND message is 5 bytes long
			size := len(data)
			if size >= 5 {
				end = data[size-5:]
			}
			switch string(end) {
			case TXEND:
				start := data[:size-5]
				if len(start) != 0 {
					addToBuffer(start)
				}
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
				addToBuffer(data)
			}
		}
	}
}
