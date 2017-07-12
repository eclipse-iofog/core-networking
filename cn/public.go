package cn

import (
	"crypto/tls"
	"time"
)

type PublicConnection struct {
	*ComSatConn
	containerConn *ContainerConn
}

func newPublicConnection(id int,
	address, remoteAddress, passcode string,
	hbInterval, hbThreshold time.Duration,
	tlsConfig *tls.Config) *PublicConnection {
	return &PublicConnection{
		ComSatConn:    newConn(id, address, passcode, hbInterval, hbThreshold, tlsConfig),
		containerConn: newContainerConn(id, remoteAddress),
	}
}

func (p *PublicConnection) Connect() {
	go p.ComSatConn.Connect()
	done := make(chan byte)
	go p.readConnection(done)
	select {
	case <-p.done:
		logger.Printf("[ PublicConnection #%d ] Stopped by demand", p.id)
		close(done)
		return
	}
}

func (p *PublicConnection) Disconnect() {
	p.ComSatConn.Disconnect()
	p.containerConn.Disconnect()
	p.done <- 0
}

func (p *PublicConnection) Reconnect() {
	p.Disconnect()
	go p.Connect()
}

func (p *PublicConnection) proxy(done <-chan byte, reconnect chan byte) {
	for {
		select {
		case <-done:
			return
		case data, ok := <-p.containerConn.out:
			if !ok {
				close(reconnect)
				return
			}
			p.in <- data
		}
	}
}

func (p *PublicConnection) readConnection(done <-chan byte) {
	reconnect := make(chan byte)
	for {
		select {
		case <-done:
			return
		case <-reconnect:
			logger.Printf("[ PublicConnection #%d ] Have to reconnect to ComSat\n", p.id)
			p.Reconnect()
			return
		case data := <-p.out:
			if !p.containerConn.isConnected {
				if err := p.containerConn.Connect(); err != nil {
					logger.Printf("[ PublicConnection #%d ] Error when connecting to container: %s\n",
						p.id, err.Error())
					continue
				} else {
					go p.containerConn.Start()
					go p.proxy(done, reconnect)
				}
			}
			p.containerConn.in <- data
		}
	}
}
