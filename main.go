package main

import (
	"errors"
	"github.com/gorilla/websocket"
	sdk "github.com/iotracks/container-sdk-go"
	"github.com/iotracks/core-networking-system-container/cn"
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
		config.Cert = "MIIETTCCAzWgAwIBAgIDAjpxMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1HZW9UcnVzdCBJbmMuMRswGQYDVQQDExJHZW9UcnVzdCBHbG9iYWwgQ0EwHhcNMTMxMjExMjM0NTUxWhcNMjIwNTIwMjM0NTUxWjBCMQswCQYDVQQGEwJVUzEWMBQGA1UEChMNR2VvVHJ1c3QgSW5jLjEbMBkGA1UEAxMSUmFwaWRTU0wgU0hBMjU2IENBMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu1jBEgEul9h9GKrIwuWF4hdsYC7JjTEFORoGmFbdVNcRjFlbPbFUrkshhTIWX1SG5tmx2GCJa1i+ctqgAEJ2sSdZTM3jutRc2aZ/uyt11UZEvexAXFm33Vmf8Wr3BvzWLxmKlRK6msrVMNI4/Bk7WxU7NtBDTdFlodSLwWBBs9ZwF8w5wJwMoD23ESJOztmpetIqYpygC04q18NhWoXdXBC5VD0tA/hJ8LySt7ecMcfpuKqCCwW5Mc0IW7siC/acjopVHHZDdvDibvDfqCl158ikh4tq8bsIyTYYZe5QQ7hdctUoOeFTPiUs2itP3YqeUFDgb5rE1RkmiQF1cwmbOwIDAQABo4IBSjCCAUYwHwYDVR0jBBgwFoAUwHqYaI2J+6sFZAwRfap9ZbjKzE4wHQYDVR0OBBYEFJfCJ1CewsnsDIgyyHyt4qYBT9pvMBIGA1UdEwEB/wQIMAYBAf8CAQAwDgYDVR0PAQH/BAQDAgEGMDYGA1UdHwQvMC0wK6ApoCeGJWh0dHA6Ly9nMS5zeW1jYi5jb20vY3Jscy9ndGdsb2JhbC5jcmwwLwYIKwYBBQUHAQEEIzAhMB8GCCsGAQUFBzABhhNodHRwOi8vZzIuc3ltY2IuY29tMEwGA1UdIARFMEMwQQYKYIZIAYb4RQEHNjAzMDEGCCsGAQUFBwIBFiVodHRwOi8vd3d3Lmdlb3RydXN0LmNvbS9yZXNvdXJjZXMvY3BzMCkGA1UdEQQiMCCkHjAcMRowGAYDVQQDExFTeW1hbnRlY1BLSS0xLTU2OTANBgkqhkiG9w0BAQsFAAOCAQEANevhiyBWlLp6vXmp9uP+bji0MsGj21hWID59xzqxZ2nVeRQb9vrsYPJ5zQoMYIp0TKOTKqDwUX/N6fmS/ZarRfViPT9gRlATPSATGC6URq7VIf5Dockj/lPEvxrYrDrK3maXI67T30pNcx9vMaJRBBZqAOv5jUOB8FChH6bKOvMoPF9RrNcKRXdLDlJiG9g4UaCSLT+Qbsh+QJ8gRhVd4FB84XavXu0R0y8TubglpK9YCa81tGJUheNI3rzSkHp6pIQNo0LyUcDUrVNlXWz4Px8G8k/Ll6BKWcZ40egDuYVtLLrhX7atKz4lecWLVtXjCYDqwSfC2Q7sRwrp0Mr82A=="
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
