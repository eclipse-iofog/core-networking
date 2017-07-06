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
	logger.Printf("[ PublicConnection #%d ] Disconnecting...", p.id)
	p.ComSatConn.Disconnect()
	p.containerConn.Disconnect()
	p.done <- 0
}

func (p *PublicConnection) Reconnect() {
	logger.Printf("[ PublicConnection #%d ] Reconnecting...", p.id)
	p.Disconnect()
	go p.Connect()
}

func (p *PublicConnection) proxy(done <-chan byte, reconnect chan byte) {
	logger.Printf("[ PublicConnection #%d ] proxy goroutine started", p.id)
	defer logger.Printf("[ PublicConnection #%d ] proxy goroutine exited", p.id)
	for {
		//logger.Printf("[ PublicConnection #%d ] proxy goroutine in cycle", p.id)
		select {
		case <-done:
			//logger.Printf("[ PublicConnection #%d ] proxy goroutine done", p.id)
			return
		case data, ok := <-p.containerConn.out:
			//logger.Printf("[ PublicConnection #%d ] proxy goroutine out", p.id)
			if !ok {
				close(reconnect)
				return
			}
			n := len(data)
			logger.Printf("[ PublicConnection #%d ] Successfully read container's intermediate: %d\n", p.id, n)
			p.in <- data
		}
	}
}

func (p *PublicConnection) readConnection(done <-chan byte) {
	//logger.Printf("[ PublicConnection #%d ] read goroutine started", p.id)
	defer logger.Printf("[ PublicConnection #%d ] read goroutine exited", p.id)
	reconnect := make(chan byte)
	for {
		select {
		case <-done:
			return
		case <-reconnect:
			logger.Printf("[ PublicConnection #%d ] Have to reconnect.", p.id)
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
			logger.Printf("[ PublicConnection #%d ] Going to send %s to container.\n", p.id, data)
			p.containerConn.in <- data
			//logger.Printf("[ PublicConnection #%d ] Has sent data to container.\n", p.id)
		}
	}
}
