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
	"github.com/eapache/channels"
	"net"
)

type ConnMonitor struct {
	id   int
	conn net.Conn

	in      *channels.RingChannel
	out     *channels.RingChannel
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
		in:   channels.NewRingChannel(channels.BufferCap(WRITE_CHANNEL_BUFFER_SIZE)),
		out:  channels.NewRingChannel(channels.BufferCap(READ_CHANNEL_BUFFER_SIZE)),
	}
}

func (m *ConnMonitor) monitor() {
	go m.write(m.err, m.done)
	go m.read(m.err, m.done)
}

func (m *ConnMonitor) write(errChannel chan<- error, done <-chan byte) {
	for {
		select {
		case <-done:
			return
		case data := <-m.in.Out():
			if _, err := m.conn.Write(data.([]byte)); err != nil {
				m.notSent = data.([]byte)
				errChannel <- err
				return
			}
		}
	}
}

func (m *ConnMonitor) read(errChannel chan<- error, done <-chan byte) {
	defer m.out.Close()
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
			m.out.In() <- temp
		}
	}
}
