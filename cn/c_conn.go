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
	if c.conn == nil {
		logger.Printf("[ ContainerConnection #%d ] Unable to start on closed connection\n", c.id)
	}
	defer func() {
		c.isConnected = false
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
	}()
	errChannel := make(chan error, 3)
	done := make(chan byte)
	defer close(done)
	c.monitor = newConnMonitor(c.id+600, c.conn, errChannel, done)
	c.monitor.monitor()
	go c.write(errChannel, done)
	go c.read(errChannel, done)
	select {
	case err := <-errChannel:
		logger.Printf("[ ContainerConnection #%d ] Error occured: %s\n", c.id, err.Error())
	case <-c.done:
		logger.Printf("[ ContainerConnection #%d ] Stopped by demand\n", c.id)
	}
}

func (c *ContainerConn) write(errChannel chan<- error, done <-chan byte) {
	for {
		select {
		case <-done:
			return
		case data := <-c.in:
			c.monitor.in <- data
		}
	}
}

func (c *ContainerConn) read(errChannel chan<- error, done <-chan byte) {
	defer close(c.out)
	for {
		select {
		case <-done:
			return
		case data, ok := <-c.monitor.out:
			if !ok {
				return
			}
			c.out <- data
		}
	}
}
