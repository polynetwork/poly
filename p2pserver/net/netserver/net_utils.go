/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The poly network is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The poly network is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with the poly network.  If not, see <http://www.gnu.org/licenses/>.
 */

package netserver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/p2pserver/common"
)

// createListener creates a net listener on the port
func createListener(port uint16) (net.Listener, error) {
	var listener net.Listener
	var err error

	isTls := config.DefConfig.P2PNode.IsTLS
	if isTls {
		listener, err = initTlsListen(port)
		if err != nil {
			log.Error("[p2p]initTlslisten failed")
			return nil, errors.New("[p2p]initTlslisten failed")
		}
	} else {
		listener, err = initNonTlsListen(port)
		if err != nil {
			log.Error("[p2p]initNonTlsListen failed")
			return nil, errors.New("[p2p]initNonTlsListen failed")
		}
	}
	return listener, nil
}

//nonTLSDial return net.Conn with nonTls
func nonTLSDial(addr string) (net.Conn, error) {
	log.Trace()
	conn, err := net.DialTimeout("tcp", addr, time.Second*common.DIAL_TIMEOUT)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//TLSDial return net.Conn with TLS
func TLSDial(nodeAddr string) (net.Conn, error) {
	CertPath := config.DefConfig.P2PNode.CertPath
	KeyPath := config.DefConfig.P2PNode.KeyPath
	CAPath := config.DefConfig.P2PNode.CAPath

	clientCertPool := x509.NewCertPool()

	cacert, err := ioutil.ReadFile(CAPath)
	if err != nil {
		log.Error("[p2p]load CA file fail", err)
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		return nil, err
	}

	ret := clientCertPool.AppendCertsFromPEM(cacert)
	if !ret {
		return nil, errors.New("[p2p]failed to parse root certificate")
	}

	conf := &tls.Config{
		RootCAs:      clientCertPool,
		Certificates: []tls.Certificate{cert},
	}

	var dialer net.Dialer
	dialer.Timeout = time.Second * common.DIAL_TIMEOUT
	conn, err := tls.DialWithDialer(&dialer, "tcp", nodeAddr, conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//initNonTlsListen return net.Listener with nonTls mode
func initNonTlsListen(port uint16) (net.Listener, error) {
	log.Trace()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		log.Error("[p2p]Error listening\n", err.Error())
		return nil, err
	}
	return listener, nil
}

//initTlsListen return net.Listener with Tls mode
func initTlsListen(port uint16) (net.Listener, error) {
	CertPath := config.DefConfig.P2PNode.CertPath
	KeyPath := config.DefConfig.P2PNode.KeyPath
	CAPath := config.DefConfig.P2PNode.CAPath

	// load cert
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		log.Error("[p2p]load keys fail", err)
		return nil, err
	}
	// load root ca
	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		log.Error("[p2p]read ca fail", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	ret := pool.AppendCertsFromPEM(caData)
	if !ret {
		return nil, errors.New("[p2p]failed to parse root certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	log.Info("[p2p]TLS listen port is ", strconv.Itoa(int(port)))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(int(port)), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
