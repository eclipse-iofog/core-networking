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
	"net"
)

type ContainerConn struct {
	id      int
	address string
	monitor *ConnMonitor
	conn    net.Conn

	in          chan []byte
	out         chan []byte
	done        chan byte
	isConnected bool
}

func newContainerConn(id int, address string) *ContainerConn {
	return &ContainerConn{
		id:      id,
		address: address,
		in:      make(chan []byte, WRITE_CHANNEL_BUFFER_SIZE),
		done:    make(chan byte),
	}
}

func (c *ContainerConn) Connect() error {
	var err error
	logger.Printf("[ ContainerConnection #%d ] Going to dial container\n", c.id)
	c.conn, err = net.Dial("tcp", c.address)
	if err != nil {
		return err
	}
	logger.Printf("[ ContainerConnection #%d ] Connected to container\n", c.id)
	c.isConnected = true
	c.out = make(chan []byte, READ_CHANNEL_BUFFER_SIZE)
	return nil
}

func (c *ContainerConn) Disconnect() {
	c.done <- 0
}

func (c *ContainerConn) Start() {
	errChannel := make(chan error, 3)
	done := make(chan byte)
	defer func() {
		c.Close()
		close(errChannel)
		close(done)
	}()
	c.monitor = newConnMonitor(c.id+600, c.conn, errChannel, done)
	c.monitor.monitor()
	go c.write(errChannel, done)
	go c.read(errChannel, done)
	select {
	case err := <-errChannel:
		logger.Printf("[ ContainerConnection #%d ] Error occured: %s\n", c.id, err.Error())
	case <-c.done:
		logger.Printf("[ ContainerConnection #%d ] Stopped on demand\n", c.id)
	}
}

func (c *ContainerConn) Close() {
	logger.Printf("[ ContainerConnection #%d ] Closing container connection\n", c.id)
	c.isConnected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *ContainerConn) write(errChannel chan<- error, done <-chan byte) {
	logger.Printf("[ ContainerConnection #%d ] Got into write of container connection\n", c.id)
	for {
		select {
		case <-done:
			logger.Printf("[ ContainerConnection #%d ] write done\n", c.id)
			return
		case data := <-c.in:
			c.monitor.in <- data
			logger.Printf("[ ContainerConnection #%d ] sent data to container pipe %s\n", c.id, data)
		}
	}
}

func (c *ContainerConn) read(errChannel chan<- error, done <-chan byte) {
	defer close(c.out)
	logger.Printf("[ ContainerConnection #%d ] Got into read of container connection\n", c.id)
	for {
		select {
		case <-done:
			logger.Printf("[ ContainerConnection #%d ] read done\n", c.id)
			return
		case data, ok := <-c.monitor.out:
			if !ok {
				logger.Printf("[ ContainerConnection #%d ] read error\n", c.id)
				return
			}
			c.out <- data
			logger.Printf("[ ContainerConnection #%d ] received data from container pipe %s\n", c.id, data)
		}
	}
}
