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
	"time"
)

type PrivateConnectionBuilder struct {
	i               int
	address         string
	passcode        string
	hbInterval      time.Duration
	hbThreshold     time.Duration
	tlsConfig       *tls.Config
	readyConnectors chan<- Connector
}

type PublicConnectionBuilder struct {
	i               int
	address         string
	remoteAddress   string
	passcode        string
	hbInterval      time.Duration
	hbThreshold     time.Duration
	tlsConfig       *tls.Config
	readyConnectors chan<- Connector
}

func newPrivateConnectionBuilder(address, passcode string, hbInterval, hbThreshold time.Duration,
	tlsConfig *tls.Config, readyConnectors chan<- Connector) ConnectorBuilder {
	return &PrivateConnectionBuilder{
		address:         address,
		passcode:        passcode,
		hbInterval:      hbInterval,
		hbThreshold:     hbThreshold,
		tlsConfig:       tlsConfig,
		readyConnectors: readyConnectors,
	}
}

func newPublicConnectionBuilder(address, remoteAddress, passcode string, hbInterval, hbThreshold time.Duration,
	tlsConfig *tls.Config) ConnectorBuilder {
	return &PublicConnectionBuilder{
		address:       address,
		remoteAddress: remoteAddress,
		passcode:      passcode,
		hbInterval:    hbInterval,
		hbThreshold:   hbThreshold,
		tlsConfig:     tlsConfig,
	}
}

func (builder *PrivateConnectionBuilder) Build() Connector {
	defer func() {
		builder.i++
	}()
	return newPrivateConnection(
		builder.i,
		builder.address,
		builder.passcode,
		builder.hbInterval,
		builder.hbThreshold,
		builder.tlsConfig,
		builder.readyConnectors,
	)
}

func (builder *PublicConnectionBuilder) Build() Connector {
	defer func() {
		builder.i++
	}()
	return newPublicConnection(
		builder.i,
		builder.address,
		builder.remoteAddress,
		builder.passcode,
		builder.hbInterval,
		builder.hbThreshold,
		builder.tlsConfig,
	)
}
