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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	sdk "github.com/ioFog/iofog-go-sdk"
	"io/ioutil"
	"time"
)

type CoreNetworking struct {
	*CNConfig
	pool        *pool
	tlsConfig   *tls.Config
	ioFogClient *sdk.IoFogClient
}

func NewCN(config *CNConfig, ioFogClient *sdk.IoFogClient) (*CoreNetworking, error) {
	certBytes, err := ioutil.ReadFile(CERT_FILE_LOCATION + CERT_FILE_NAME)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certBytes)
	return &CoreNetworking{
		pool:        newPool(config.ConnectionCount),
		CNConfig:    config,
		ioFogClient: ioFogClient,
		tlsConfig: &tls.Config{
			RootCAs: certPool,
		},
	}, nil
}

func (cn *CoreNetworking) Start() {
	switch cn.Mode {
	case MODE_PUBLIC:
		builder := newPublicConnectionBuilder(
			fmt.Sprint(cn.Host, ":", cn.Port),
			fmt.Sprint(cn.LocalHost, ":", cn.LocalPort),
			cn.Passcode,
			time.Millisecond*time.Duration(cn.HBFrequency),
			time.Millisecond*time.Duration(cn.HBThreshold),
			cn.tlsConfig,
		)
		cn.pool.start(builder)
	case MODE_PRIVATE:
		builder := newPrivateConnectionBuilder(
			fmt.Sprint(cn.Host, ":", cn.Port),
			cn.Passcode,
			time.Millisecond*time.Duration(cn.HBFrequency),
			time.Millisecond*time.Duration(cn.HBThreshold),
			cn.tlsConfig,
			cn.pool.readyConnectors,
		)
		cn.pool.start(builder)
		dataChannel, receiptChannel := cn.ioFogClient.EstablishMessageWsConnection(0, 0)
		go func() {
			<-receiptChannel
		}()
		go cn.pool.sendMessagesFromBus(dataChannel)
		go cn.pool.sendMessagesToBus(cn.ioFogClient)
	case MODE_P2P:
	}
}

func (cn *CoreNetworking) Stop() {
	cn.pool.stop()
}
