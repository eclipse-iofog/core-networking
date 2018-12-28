/********************************************************************************
 * Copyright (c) 2018 Edgeworx, Inc.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0 which is available at
 * http://www.eclipse.org/legal/epl-2.0
 *
 * SPDX-License-Identifier: EPL-2.0
 ********************************************************************************/

package cn

import (
	"crypto/tls"
	"github.com/eapache/channels"
	sdk "github.com/ioFog/iofog-go-sdk"
	"time"
)

type PrivateConnection struct {
	*ConnectorConn

	inMessage  *channels.RingChannel
	outMessage *channels.RingChannel
	readyConn  chan<- Connector
}

func newPrivateConnection(id int,
	address, passcode string,
	hbInterval, hbThreshold time.Duration,
	tlsConfig *tls.Config, devMode bool,
	ready chan<- Connector) *PrivateConnection {
	return &PrivateConnection{
		ConnectorConn: newConn(id, address, passcode, hbInterval, hbThreshold, tlsConfig, devMode),
		inMessage:     channels.NewRingChannel(channels.BufferCap(READ_CHANNEL_BUFFER_SIZE)),
		outMessage:    channels.NewRingChannel(channels.BufferCap(WRITE_CHANNEL_BUFFER_SIZE)),
		readyConn:     ready,
	}
}

func (p *PrivateConnection) Connect() {
	go p.ConnectorConn.Connect()
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
	p.ConnectorConn.Disconnect()
	p.done <- 0
}

func (p *PrivateConnection) writeConnection(done <-chan byte) {
	for {
		select {
		case msg := <-p.inMessage.Out():
			if bytes, err := sdk.PrepareMessageForSendingViaSocket(msg.(*sdk.IoMessage)); err != nil {
				logger.Printf("[ PrivateConnection #%d ] Error while encoding message: %s\n", p.id, err.Error())
			} else {
				p.in.In() <- bytes
				p.in.In() <- []byte(TXEND)
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
		case data := <-p.out.Out():
			dataArr := data.([]byte)
			end := make([]byte, 5) // TXEND message is 5 bytes long
			size := len(dataArr)
			if size >= 5 {
				end = dataArr[size-5:]
			}
			switch string(end) {
			case TXEND:
				start := dataArr[:size-5]
				if len(start) != 0 {
					addToBuffer(start)
				}
				if !isBroken {
					if msg, err := sdk.GetMessageReceivedViaSocket(b); err != nil {
						logger.Printf("[ PrivateConnection #%d ] Error while decoding message: %s", p.id, err.Error())
					} else {
						p.outMessage.In() <- msg
					}
				}
				b = b[:0]
				isBroken = false
			default:
				addToBuffer(dataArr)
			}
		}
	}
}
