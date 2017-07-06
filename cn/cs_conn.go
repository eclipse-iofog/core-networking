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

	in          chan []byte
	out         chan []byte
	done        chan byte
	isConnected bool
	notSent     []byte

	monitor  *ConnMonitor
	latMutex sync.Mutex
}

func newConn(id int, address, passcode string, hbInterval, hbThreshold time.Duration, tlsConfig *tls.Config) *ComSatConn {
	return &ComSatConn{
		id:          id,
		address:     address,
		passcode:    passcode,
		hbInterval:  hbInterval,
		hbThreshold: hbThreshold,
		tlsConfig:   tlsConfig,
		in:          make(chan []byte, WRITE_CHANNEL_BUFFER_SIZE),
		out:         make(chan []byte, READ_CHANNEL_BUFFER_SIZE),
		done:        make(chan byte),
	}
}

func (c *ComSatConn) Connect() {
	logger.Printf("[ Connection #%d ] started\n", c.id)
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
			logger.Printf("[ Connection #%d ] stopped by demand\n", c.id)
			return
		default:
			logger.Printf("[ Connection #%d ] before dial %s\n", c.id, c.address)
			conn, err = tls.Dial("tcp", c.address, c.tlsConfig)
			logger.Printf("[ Connection #%d ] after dial %s\n", c.id, c.address)
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
					logger.Printf("[ Connection #%d ] Retrying to send data\n", c.id)
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
				c.monitor.Monitor()
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
					logger.Printf("[ Connection #%d ] Stopped by demand\n", c.id)
					close(done)
					return
				}
				c.isConnected = false
				conn.Close()
				c.notSent = c.monitor.notSent
				logger.Printf("[ Connection #%d ] Will send on next connect: %s\n", c.id, c.notSent)
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
	defer logger.Printf("[ Connection #%d ] lat monitor goroutine exited\n", c.id)
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
	defer logger.Printf("[ Connection #%d ] write goroutine exited\n", c.id)
	hbTicker := time.NewTicker(c.hbInterval)
	defer hbTicker.Stop()
	for {
		select {
		case <-done:
			return
		case <-hbTicker.C:
			c.monitor.in <- []byte(BEAT)
		case data := <-c.in:
			c.monitor.in <- data
		}
	}
}

func (c *ComSatConn) read(errChannel chan<- error, done <-chan byte) {
	defer logger.Printf("[ Connection #%d ] read goroutine exited\n", c.id)
	for {
		select {
		case <-done:
			return
		case data, ok := <-c.monitor.out:
			if !ok {
				return
			}
			c.latMutex.Lock()
			c.lastActivityTime = time.Now()
			c.latMutex.Unlock()
			switch string(data) {
			case BEAT:
			case DOUBLE_BEAT:
			default:
				logger.Printf("[ Connection #%d ] has read %s\n", c.id, data)
				c.out <- data
			}
		}
	}
}
