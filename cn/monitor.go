package cn

import (
	"net"
)

type ConnMonitor struct {
	id   int
	conn net.Conn

	in      chan []byte
	out     chan []byte
	err     chan<- error
	done    chan byte
	notSent []byte
}

func newConnMonitor(id int, conn net.Conn, err chan<- error, done chan byte) *ConnMonitor {
	return &ConnMonitor{
		id:   id,
		conn: conn,
		err:  err,
		done: done,
		in:   make(chan []byte, WRITE_CHANNEL_BUFFER_SIZE),
		out:  make(chan []byte, READ_CHANNEL_BUFFER_SIZE),
	}
}

func (m *ConnMonitor) Monitor() {
	go m.write(m.err, m.done)
	go m.read(m.err, m.done)
}

func (m *ConnMonitor) write(errChannel chan<- error, done <-chan byte) {
	logger.Printf("[ Monitoring #%d ] write goroutine started\n", m.id)
	defer logger.Printf("[ Monitoring #%d ] write goroutine exited\n", m.id)
	for {
		select {
		case <-done:
			return
		case data := <-m.in:
			if n, err := m.conn.Write(data); err != nil {
				m.notSent = data
				errChannel <- err
				return
			} else {
				logger.Printf("[ Monitoring #%d ] Has wrote %d bytes\n", m.id, n)
			}
		}
	}
}

func (m *ConnMonitor) read(errChannel chan<- error, done <-chan byte) {
	logger.Printf("[ Monitoring #%d ] read goroutine started\n", m.id)
	defer logger.Printf("[ Monitoring #%d ] read goroutine exited\n", m.id)
	defer close(m.out)
	p := make([]byte, DEFAULT_READ_SIZE)
	for {
		select {
		case <-done:
			return
		default:
			n, err := m.conn.Read(p)
			if err != nil {
				errChannel <- err
				return
			}
			temp := make([]byte, n)
			copy(temp, p)
			m.out <- temp
		}
	}
}
