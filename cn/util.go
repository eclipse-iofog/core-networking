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
	"log"
	"os"
	"time"
)

const (
	MODE_PRIVATE = "private"
	MODE_PUBLIC  = "public"
	MODE_P2P     = "peer-to-peer"

	BEAT        = "BEAT"
	DOUBLE_BEAT = "BEATBEAT"
	TXEND       = "TXEND"
	AUTHORIZED  = "AUTHORIZED"

	CERT_FILE_NAME     = "cert.pem"
	CERT_FILE_LOCATION = "./"

	ATTEMPT_LIMIT      = 10
	CONNECT_TIMEOUT    = time.Second

	WRITE_CHANNEL_BUFFER_SIZE = 20
	READ_CHANNEL_BUFFER_SIZE  = 20

	MAX_READ_BUFFER_SIZE = 10 * 1024 * 1024 // 10MB
	DEFAULT_READ_SIZE    = 64 * 1024        // 64KB

)

var logger = log.New(os.Stderr, "", log.LstdFlags)

func WriteCertToFile(cert string) error {
	file, err := os.Create(CERT_FILE_LOCATION + CERT_FILE_NAME)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("-----BEGIN CERTIFICATE-----\n")
	end := 64
	for len(cert) > 0 {
		if len(cert) < 64 {
			end = len(cert)
		}
		file.WriteString(cert[:end])
		file.WriteString("\n")
		cert = cert[end:]
	}
	file.WriteString("-----END CERTIFICATE-----\n")
	return nil
}

type CNConfig struct {
	Mode            string `json:"mode"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	ConnectionCount int    `json:"connectioncount"`
	Passcode        string `json:"passcode"`
	LocalHost       string `json:"localhost"`
	LocalPort       int    `json:"localport"`
	HBFrequency     uint   `json:"heartbeatfrequency"`
	HBThreshold     uint   `json:"heartbeatabsencethreshold"`
	Cert            string `json:"cert"`
	DevMode		    bool   `json:"devmode"`
}
