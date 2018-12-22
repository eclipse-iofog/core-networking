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
	sdk "github.com/ioFog/iofog-go-sdk"
)

type pool struct {
	Connectors      []Connector
	Count           int
	readyConnectors chan Connector
}

func newPool(connCount int) *pool {
	return &pool{
		Count:           connCount,
		Connectors:      make([]Connector, connCount),
		readyConnectors: make(chan Connector, connCount),
	}
}

func (pool *pool) start(connectorBuilder ConnectorBuilder) {
	for i := 0; i < pool.Count; i++ {
		pool.Connectors[i] = connectorBuilder.Build()
		go pool.Connectors[i].Connect()
	}
}

func (pool *pool) messagesFromComSat() <-chan interface{} {
	out := channels.NewRingChannel(channels.BufferCap(READ_CHANNEL_BUFFER_SIZE*pool.Count))
	output := func(c <-chan interface{}) {
		for n := range c {
			out.In() <- n
		}
	}
	for _, c := range pool.Connectors {
		go output(c.(*PrivateConnection).outMessage.Out())
	}

	return out.Out()
}

func (pool *pool) sendMessagesFromBus(incomingMessages <-chan interface{}) {
	for msg := range incomingMessages {
		c := <-pool.readyConnectors
		c.(*PrivateConnection).inMessage.In() <- msg
	}
}
func (pool *pool) sendMessagesToBus(ioFogClient *sdk.IoFogClient) {
	for msg := range pool.messagesFromComSat() {
		ioFogClient.SendMessageViaSocket(msg.(*sdk.IoMessage))
	}
}

func (pool *pool) stop() {
	for _, c := range pool.Connectors {
		c.Disconnect()
	}
}
