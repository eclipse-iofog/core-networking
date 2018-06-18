/********************************************************************************
 * Copyright (c) 2018 Edgeworx, Inc.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0 which is available at
 * http://www.eclipse.org/legal/epl-2.0
 *
 * SPDX-License-Identifier: EPL-2.0
 ********************************************************************************/

package main

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/ioFog/core-networking/cn"
	sdk "github.com/ioFog/iofog-go-sdk"
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
		config.Cert = "MIIEsTCCA5mgAwIBAgIQCKWiRs1LXIyD1wK0u6tTSTANBgkqhkiG9w0BAQsFADBhMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBDQTAeFw0xNzExMDYxMjIzMzNaFw0yNzExMDYxMjIzMzNaMF4xCzAJBgNVBAYTAlVTMRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5jb20xHTAbBgNVBAMTFFJhcGlkU1NMIFJTQSBDQSAyMDE4MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5S2oihEo9nnpezoziDtx4WWLLCll/e0t1EYemE5n+MgP5viaHLy+VpHP+ndX5D18INIuuAV8wFq26KF5U0WNIZiQp6mLtIWjUeWDPA28OeyhTlj9TLk2beytbtFU6ypbpWUltmvY5V8ngspC7nFRNCjpfnDED2kRyJzO8yoKMFz4J4JE8N7NA1uJwUEFMUvHLs0scLoPZkKcewIRm1RV2AxmFQxJkdf7YN9Pckkif2Xgm3b48BZn0zf0qXsSeGu84ua9gwzjzI7tbTBjayTpT+/XpWuBVv6fvarI6bikKB859OSGQuw73XXgeuFwEPHTIRoUtkzu3/EQ+LtwznkkdQIDAQABo4IBZjCCAWIwHQYDVR0OBBYEFFPKF1n8a8ADIS8aruSqqByCVtp1MB8GA1UdIwQYMBaAFAPeUDVW0Uy7ZvCj4hsbw5eyPdFVMA4GA1UdDwEB/wQEAwIBhjAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwEgYDVR0TAQH/BAgwBgEB/wIBADA0BggrBgEFBQcBAQQoMCYwJAYIKwYBBQUHMAGGGGh0dHA6Ly9vY3NwLmRpZ2ljZXJ0LmNvbTBCBgNVHR8EOzA5MDegNaAzhjFodHRwOi8vY3JsMy5kaWdpY2VydC5jb20vRGlnaUNlcnRHbG9iYWxSb290Q0EuY3JsMGMGA1UdIARcMFowNwYJYIZIAYb9bAECMCowKAYIKwYBBQUHAgEWHGh0dHBzOi8vd3d3LmRpZ2ljZXJ0LmNvbS9DUFMwCwYJYIZIAYb9bAEBMAgGBmeBDAECATAIBgZngQwBAgIwDQYJKoZIhvcNAQELBQADggEBAH4jx/LKNW5ZklFcYWs8Ejbm0nyzKeZC2KOVYR7P8gevKyslWm4Xo4BSzKr235FsJ4aFt6yAiv1eY0tZ/ZN18bOGSGStoEc/JE4ocIzr8P5Mg11kRYHbmgYnr1Rxeki5mSeb39DGxTpJD4kGhs5lXNoo4conUiiJwKaqH7vh2baryd8pMISag83JUqyVGc2tWPpO0329/CWq2kryqv66OSMjwulUz0dXf4OHQasR7CNfIr+4KScc6ABlQ5RDF86PGeE6kdwSQkFiB/cQysNyq0jEDQTkfa2pjmuWtMCNbBnhFXBYejfubIhaUbEv2FOQB3dCav+FPg5eEveXTVyMnGo="
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
