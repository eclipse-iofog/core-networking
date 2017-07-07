package main

import (
	"errors"
	"github.com/gorilla/websocket"
	sdk "github.com/iotracks/container-sdk-go"
	"log"
	"os"
	"time"
	"github.com/iotracks/core-networking-system-container/cn"
)

var (
	logger = log.New(os.Stderr, "", log.LstdFlags)
	ioFogClient, clientError = sdk.NewDefaultIoFogClient()
)

type ConnectionManager struct {
	connection  *websocket.Conn
	lastUpdated time.Time
}

func main() {

	if clientError != nil {
		logger.Fatalln(clientError.Error())
	}

	config := new(cn.CNConfig)
	if err := updateConfig(config); err != nil {
		logger.Fatalln(clientError.Error())
	}


	cn.WriteCertToFile(config.Cert)
	c, err := cn.NewCN(config, ioFogClient)
	if err != nil {
		logger.Fatalln(err.Error())
	}
	c.Start()

	for {
		time.Sleep(time.Second)
	}

	confChannel := ioFogClient.EstablishControlWsConnection(0)
	for {
		select {
		case <-confChannel:
			logger.Println("[ Main ] Config update is not supported yet")
			//updateConfig(config)
		}
	}
}

func updateConfig(config interface{}) error {
	attemptLimit := 5
	var err error

	for err = ioFogClient.GetConfigIntoStruct(config); err != nil && attemptLimit > 0; attemptLimit-- {
		logger.Println(err.Error())
	}

	if attemptLimit == 0 {
		return errors.New("Update config failed")
	}

	return nil
}
