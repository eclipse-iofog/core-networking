package main

import (
	"errors"
	"github.com/gorilla/websocket"
	sdk "github.com/ioFog/iofog-go-sdk"
	"github.com/ioFog/core-networking/cn"
	"log"
	"os"
	"time"
)

var (
	logger                   = log.New(os.Stderr, "", log.LstdFlags)
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
		logger.Fatalln(err.Error())
	}

	if config.Cert == "" {
		config.Cert = "MIIDUDCCAjigAwIBAgIQFtoPZPmXdsWuZJtSRbGNDTANBgkqhkiG9w0BAQsFADA0MRQwEgYDVQQKEwtMb2cgQ291cmllcjEcMBoGA1UEAxMTZWRnZXdvcmtzLmxvY2FsLmNvbTAeFw0xODA0MDIxNTAwMTBaFw0xOTA4MTUxNTAwMTBaMDQxFDASBgNVBAoTC0xvZyBDb3VyaWVyMRwwGgYDVQQDExNlZGdld29ya3MubG9jYWwuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwvEpFn37F94sXrrs8p2b3i5RlwjfTBFDe1EiGQsDFB4bLeon2ME7ZyDFgb6Ol9F0b0dEr2NZ6mn9OI7xjA3goMpdxBJ8rrCL2uz1TSAgvape2AE0l/bQIkYdpUcdLAelu382c9sfCr30jsp6lD1Zuqe1MrbQQbX3iLcQFsYR6fU929dx1k+fMaETVmwBLOvTeFbyoFOmzj4oOzp6w8C0EEZMlU/f9n0exIsLDfrVcQSvwr/5dbNR0IBtLc+BAiwGSufN/1ucC4syHaSnoNxLs9C6cOQStHqyAD6uJrDOXxz7dzuEXCSh50Xz5eSDbhf0ITLNITY/n1MNcDOd8RjaiwIDAQABo14wXDAOBgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUwAwEB/zAkBgNVHREEHTAbghNlZGdld29ya3MubG9jYWwuY29thwTAqAHPMA0GCSqGSIb3DQEBCwUAA4IBAQBHg4hdFT+IIxOA2/HgDjW4u3VrNGQFkNvMBOJDVK9fA+l/cxA8bb1Btf7dv/utQTBvGF3a9FKru+JXcccqLpFZpyfcGAfnrMQ8AAaS3M0xzMcBKtkrOCYdGB7T286sC7QKCbenooFeDNo6m93Gg5WKJnXmoHLK7OHC9NkyXkBCz+a3k5xfHGgg3ygbLcCyQU7hhVuUs2HbinsvS6r4qni71fmwUIWWc0T12yB6PjNSTQLDxYV8efHE5ST4neQE2/g/7+UhKXghwUJoTUVkTPKTqL1MuB609/lbCoCZMZymSAWkeNAUgenG6JoGECyJ6ubZ0o6JYA25fZ1v14QbVBEt"
	}
	cn.WriteCertToFile(config.Cert)
	c, err := cn.NewCN(config, ioFogClient)
	if err != nil {
		logger.Fatalln(err.Error())
	}
	c.Start()

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
