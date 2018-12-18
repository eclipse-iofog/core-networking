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
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type ComSatConn struct {
	id               int
	address          string
	passcode         string
	lastActivityTime time.Time
	hbInterval       time.Duration
	hbThreshold      time.Duration
	tlsConfig        *tls.Config
	devMode			 bool
	in          chan []byte
	out         chan []byte
	done        chan byte
	isConnected bool
	notSent     []byte

	monitor  *ConnMonitor
	latMutex sync.Mutex
}

func newConn(id int, address, passcode string, hbInterval, hbThreshold time.Duration, tlsConfig *tls.Config, devMode bool) *ComSatConn {
	return &ComSatConn{
		id:          id,
		address:     address,
		passcode:    passcode,
		hbInterval:  hbInterval,
		hbThreshold: hbThreshold,
		tlsConfig:   tlsConfig,
		devMode:	 devMode,
		in:          make(chan []byte, WRITE_CHANNEL_BUFFER_SIZE),
		out:         make(chan []byte, READ_CHANNEL_BUFFER_SIZE),
		done:        make(chan byte),
	}
}

func (c *ComSatConn) Connect() {
	var conn net.Conn
	var err error
	defer func() {
		c.isConnected = false
		if conn != nil {
			conn.Close()
		}
	}()
	attempt := uint(0)
	for {
		select {
		case <-c.done:
			logger.Printf("[ Connection #%d ] stopped on demand\n", c.id)
			return
		default:
			if c.devMode {
				logger.Printf("[ Connection #%d ] Going to dial ComSat in dev mode\n", c.id)
				conn, err = net.Dial("tcp", c.address)
			} else {
				logger.Printf("[ Connection #%d ] Going to dial ComSat\n in regular mode", c.id)
				conn, err = tls.Dial("tcp", c.address, c.tlsConfig)
			}
			if err != nil {
				sleepTime := 1 << attempt * CONNECT_TIMEOUT
				if attempt < ATTEMPT_LIMIT {
					attempt++
				}
				logger.Printf("[ Connection #%d ] Error when dialing ComSat: %s. Retrying after %v\n",
					c.id, err.Error(), sleepTime)
				time.Sleep(sleepTime)
			} else {
				attempt = 0
				logger.Printf("[ Connection #%d ] Connected to ComSat\n", c.id)
				if err := c.authorize(conn); err != nil {
					logger.Printf("[ Connection #%d ] Error while authorizing: %s\n", c.id, err.Error())
				}
				if c.notSent != nil {
					if _, err := conn.Write(c.notSent); err != nil {
						logger.Printf("[ Connection #%d ] Error when retrying to send data: %s\n",
							c.id, err.Error())
						continue
					}
					logger.Printf("[ Connection #%d ] Successfully sent data on retry\n", c.id)
					c.notSent = nil
				}
				errChannel := make(chan error, 3)
				done := make(chan byte)
				c.monitor = newConnMonitor(c.id, conn, errChannel, done)
				c.monitor.monitor()
				c.isConnected = true
				c.lastActivityTime = time.Now()
				go c.write(errChannel, done)
				go c.read(errChannel, done)
				go c.monitorLastActivityTime(errChannel, done)
				select {
				case err := <-errChannel:
					logger.Printf("[ Connection #%d ] Error occured: %s\n", c.id, err.Error())
					close(done)
				case <-c.done:
					logger.Printf("[ Connection #%d ] Stopped on demand\n", c.id)
					close(done)
					return
				}
				c.isConnected = false
				conn.Close()
				c.notSent = c.monitor.notSent
			}
		}
	}
}

func (c *ComSatConn) Disconnect() {
	c.done <- 0
}

func (c *ComSatConn) authorize(conn net.Conn) error {
	if _, err := conn.Write([]byte(c.passcode)); err != nil {
		return errors.New(fmt.Sprintf("Error while sending passcode: %s", err.Error()))
	}
	p := make([]byte, len(AUTHORIZED))
	if _, err := conn.Read(p); err != nil {
		return errors.New(fmt.Sprintf("Error while reading %s: %s", AUTHORIZED, err.Error()))
	}
	if string(p) != AUTHORIZED {
		return errors.New(fmt.Sprintf("Did not receive '%s'", AUTHORIZED))
	}
	return nil
}

func (c *ComSatConn) monitorLastActivityTime(errChannel chan<- error, done <-chan byte) {
	hbTicker := time.NewTicker(c.hbThreshold)
	defer hbTicker.Stop()
	for {
		select {
		case <-done:
			return
		case t := <-hbTicker.C:
			c.latMutex.Lock()
			sub := t.Sub(c.lastActivityTime)
			c.latMutex.Unlock()
			if sub >= c.hbThreshold {
				errChannel <- errors.New("Heartbeat threshold triggered\n")
				return
			}
		}
	}
}

func (c *ComSatConn) write(errChannel chan<- error, done <-chan byte) {
	hbTicker := time.NewTicker(c.hbInterval)
	defer hbTicker.Stop()
	logger.Printf("[ Connection #%d ] got into write of comsat connection\n", c.id)
	for {
		select {
		case <-done:
			logger.Printf("[ Connection #%d ] write done\n", c.id)
			return
		case <-hbTicker.C:
			c.monitor.in <- []byte(BEAT)
		case data := <-c.in:
			c.monitor.in <- data
			logger.Printf("[ Connection #%d ] sent data to comsat %s\n", c.id, data)
		}
	}
}

func (c *ComSatConn) read(errChannel chan<- error, done <-chan byte) {
	logger.Printf("[ Connection #%d ] got into read of comsat connection\n", c.id)
	for {
		select {
		case <-done:
			logger.Printf("[ Connection #%d ] read done\n", c.id)
			return
		case data, ok := <-c.monitor.out:
			if !ok {
				logger.Printf("[ Connection #%d ] read error\n", c.id)
				return
			}
			c.latMutex.Lock()
			c.lastActivityTime = time.Now()
			c.latMutex.Unlock()
			switch string(data) {
			case BEAT:
			case DOUBLE_BEAT:
			default:
				c.out <- data
				logger.Printf("[ Connection #%d ] received data from comsat %s\n", c.id, data)
			}
		}
	}
}
