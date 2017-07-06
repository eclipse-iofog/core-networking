package cn

type Connector interface {
	Connect()
	Disconnect()
}

type ConnectorBuilder interface {
	Build() Connector
}
