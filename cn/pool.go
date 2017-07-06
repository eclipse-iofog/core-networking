package cn

import (
	sdk "github.com/iotracks/container-sdk-go"
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

func (pool *pool) messagesFromComSat() <-chan *sdk.IoMessage {
	out := make(chan *sdk.IoMessage, READ_CHANNEL_BUFFER_SIZE*pool.Count)
	output := func(c <-chan *sdk.IoMessage) {
		for n := range c {
			out <- n
		}
	}
	for _, c := range pool.Connectors {
		go output(c.(*PrivateConnection).outMessage)
	}

	return out
}

func (pool *pool) sendMessagesFromBus(incomingMessages <-chan *sdk.IoMessage) {
	for msg := range incomingMessages {
		logger.Println("[ Pool ] Waiting for ready connector")
		c := <-pool.readyConnectors
		logger.Println("[ Pool ] Got one ready connector")
		c.(*PrivateConnection).inMessage <- msg
	}
}
func (pool *pool) sendMessagesToBus(ioFogClient *sdk.IoFogClient) {
	for msg := range pool.messagesFromComSat() {
		logger.Printf("[ Pool ] Writing to message socket %v\n", msg)
		ioFogClient.SendMessageViaSocket(msg)
	}
}

func (pool *pool) stop() {
	for _, c := range pool.Connectors {
		c.Disconnect()
	}
}
