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

package peer

import (
	"errors"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	comm "github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/p2pserver/common"
	conn "github.com/polynetwork/poly/p2pserver/link"
	"github.com/polynetwork/poly/p2pserver/message/types"
)

// PeerCom provides the basic information of a peer
type PeerCom struct {
	id           uint64
	version      uint32
	services     uint64
	relay        bool
	httpInfoPort uint16
	syncPort     uint16
	consPort     uint16
	height       uint64
	softVersion  string
}

// SetID sets a peer's id
func (this *PeerCom) SetID(id uint64) {
	this.id = id
}

// GetID returns a peer's id
func (this *PeerCom) GetID() uint64 {
	return this.id
}

// SetVersion sets a peer's version
func (this *PeerCom) SetVersion(version uint32) {
	this.version = version
}

// GetVersion returns a peer's version
func (this *PeerCom) GetVersion() uint32 {
	return this.version
}

// SetServices sets a peer's services
func (this *PeerCom) SetServices(services uint64) {
	this.services = services
}

// GetServices returns a peer's services
func (this *PeerCom) GetServices() uint64 {
	return this.services
}

// SerRelay sets a peer's relay
func (this *PeerCom) SetRelay(relay bool) {
	this.relay = relay
}

// GetRelay returns a peer's relay
func (this *PeerCom) GetRelay() bool {
	return this.relay
}

// SetSyncPort sets a peer's sync port
func (this *PeerCom) SetSyncPort(port uint16) {
	this.syncPort = port
}

// GetSyncPort returns a peer's sync port
func (this *PeerCom) GetSyncPort() uint16 {
	return this.syncPort
}

// SetConsPort sets a peer's consensus port
func (this *PeerCom) SetConsPort(port uint16) {
	this.consPort = port
}

// GetConsPort returns a peer's consensus port
func (this *PeerCom) GetConsPort() uint16 {
	return this.consPort
}

// SetHttpInfoPort sets a peer's http info port
func (this *PeerCom) SetHttpInfoPort(port uint16) {
	this.httpInfoPort = port
}

// GetHttpInfoPort returns a peer's http info port
func (this *PeerCom) GetHttpInfoPort() uint16 {
	return this.httpInfoPort
}

// SetHeight sets a peer's height
func (this *PeerCom) SetHeight(height uint64) {
	this.height = height
}

// GetHeight returns a peer's height
func (this *PeerCom) GetHeight() uint64 {
	return this.height
}

//SetSoftVersion sets a peers's software version
func (this *PeerCom) SetSoftVersion(softVer string) {
	this.softVersion = softVer
}

//GetSoftVersion return a peer's software version
func (this *PeerCom) GetSoftVersion() string {
	return this.softVersion
}

//Peer represent the node in p2p
type Peer struct {
	base      PeerCom
	cap       [32]byte
	SyncLink  *conn.Link
	ConsLink  *conn.Link
	syncState uint32
	consState uint32
	txnCnt    uint64
	rxTxnCnt  uint64
	connLock  sync.RWMutex
}

//NewPeer return new peer without publickey initial
func NewPeer() *Peer {
	p := &Peer{
		syncState: common.INIT,
		consState: common.INIT,
	}
	p.SyncLink = conn.NewLink()
	p.ConsLink = conn.NewLink()
	runtime.SetFinalizer(p, rmPeer)
	return p
}

//rmPeer print a debug log when peer be finalized by system
func rmPeer(p *Peer) {
	log.Debugf("[p2p]Remove unused peer: %d", p.GetID())
}

//DumpInfo print all information of peer
func (this *Peer) DumpInfo() {
	log.Debug("[p2p]Node info:")
	log.Debug("[p2p]\t syncState = ", this.syncState)
	log.Debug("[p2p]\t consState = ", this.consState)
	log.Debug("[p2p]\t id = ", this.GetID())
	log.Debug("[p2p]\t addr = ", this.SyncLink.GetAddr())
	log.Debug("[p2p]\t cap = ", this.cap)
	log.Debug("[p2p]\t version = ", this.GetVersion())
	log.Debug("[p2p]\t services = ", this.GetServices())
	log.Debug("[p2p]\t syncPort = ", this.GetSyncPort())
	log.Debug("[p2p]\t consPort = ", this.GetConsPort())
	log.Debug("[p2p]\t relay = ", this.GetRelay())
	log.Debug("[p2p]\t height = ", this.GetHeight())
	log.Debug("[p2p]\t softVersion = ", this.GetSoftVersion())
}

//GetVersion return peer`s version
func (this *Peer) GetVersion() uint32 {
	return this.base.GetVersion()
}

//GetHeight return peer`s block height
func (this *Peer) GetHeight() uint64 {
	return this.base.GetHeight()
}

//SetHeight set height to peer
func (this *Peer) SetHeight(height uint64) {
	this.base.SetHeight(height)
}

//GetConsConn return consensus link
func (this *Peer) GetConsConn() *conn.Link {
	return this.ConsLink
}

//SetConsConn set consensue link to peer
func (this *Peer) SetConsConn(consLink *conn.Link) {
	this.ConsLink = consLink
}

//GetSyncState return sync state
func (this *Peer) GetSyncState() uint32 {
	return this.syncState
}

//SetSyncState set sync state to peer
func (this *Peer) SetSyncState(state uint32) {
	atomic.StoreUint32(&(this.syncState), state)
}

//GetConsState return peer`s consensus state
func (this *Peer) GetConsState() uint32 {
	return this.consState
}

//SetConsState set consensus state to peer
func (this *Peer) SetConsState(state uint32) {
	atomic.StoreUint32(&(this.consState), state)
}

//GetSyncPort return peer`s sync port
func (this *Peer) GetSyncPort() uint16 {
	return this.SyncLink.GetPort()
}

//GetConsPort return peer`s consensus port
func (this *Peer) GetConsPort() uint16 {
	return this.ConsLink.GetPort()
}

//SetConsPort set peer`s consensus port
func (this *Peer) SetConsPort(port uint16) {
	this.ConsLink.SetPort(port)
}

//SendToSync call sync link to send buffer
func (this *Peer) SendToSync(msgType string, msgPayload []byte) error {
	if this.SyncLink != nil && this.SyncLink.Valid() {
		return this.SyncLink.SendRaw(msgPayload)
	}
	return errors.New("[p2p]sync link invalid")
}

//SendToCons call consensus link to send buffer
func (this *Peer) SendToCons(msgType string, msgPayload []byte) error {
	if this.ConsLink != nil && this.ConsLink.Valid() {
		return this.ConsLink.SendRaw(msgPayload)
	}
	return errors.New("[p2p]cons link invalid")
}

//CloseSync halt sync connection
func (this *Peer) CloseSync() {
	this.SetSyncState(common.INACTIVITY)
	conn := this.SyncLink.GetConn()
	this.connLock.Lock()
	if conn != nil {
		conn.Close()
	}
	this.connLock.Unlock()
}

//CloseCons halt consensus connection
func (this *Peer) CloseCons() {
	this.SetConsState(common.INACTIVITY)
	conn := this.ConsLink.GetConn()
	this.connLock.Lock()
	if conn != nil {
		conn.Close()

	}
	this.connLock.Unlock()
}

//GetID return peer`s id
func (this *Peer) GetID() uint64 {
	return this.base.GetID()
}

//GetRelay return peer`s relay state
func (this *Peer) GetRelay() bool {
	return this.base.GetRelay()
}

//GetServices return peer`s service state
func (this *Peer) GetServices() uint64 {
	return this.base.GetServices()
}

//GetTimeStamp return peer`s latest contact time in ticks
func (this *Peer) GetTimeStamp() int64 {
	return this.SyncLink.GetRXTime().UnixNano()
}

//GetContactTime return peer`s latest contact time in Time struct
func (this *Peer) GetContactTime() time.Time {
	return this.SyncLink.GetRXTime()
}

//GetAddr return peer`s sync link address
func (this *Peer) GetAddr() string {
	return this.SyncLink.GetAddr()
}

//GetAddr16 return peer`s sync link address in []byte
func (this *Peer) GetAddr16() ([16]byte, error) {
	var result [16]byte
	addrIp, err := common.ParseIPAddr(this.GetAddr())
	if err != nil {
		return result, err
	}
	ip := net.ParseIP(addrIp).To16()
	if ip == nil {
		log.Warn("[p2p]parse ip address error\n", this.GetAddr())
		return result, errors.New("[p2p]parse ip address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

func (this *Peer) GetSoftVersion() string {
	return this.base.GetSoftVersion()
}

//AttachSyncChan set msg chan to sync link
func (this *Peer) AttachSyncChan(msgchan chan *types.MsgPayload) {
	this.SyncLink.SetChan(msgchan)
}

//AttachConsChan set msg chan to consensus link
func (this *Peer) AttachConsChan(msgchan chan *types.MsgPayload) {
	this.ConsLink.SetChan(msgchan)
}

//Send transfer buffer by sync or cons link
func (this *Peer) Send(msg types.Message, isConsensus bool) error {
	sink := comm.NewZeroCopySink(nil)
	err := types.WriteMessage(sink, msg)
	if err != nil {
		log.Debugf("[p2p]error serialize messge ", err.Error())
		return err
	}

	return this.SendRaw(msg.CmdType(), sink.Bytes(), isConsensus)
}

func (this *Peer) SendRaw(msgType string, msgPayload []byte, isConsensus bool) error {
	if isConsensus && this.ConsLink.Valid() {
		return this.SendToCons(msgType, msgPayload)
	}
	return this.SendToSync(msgType, msgPayload)
}

//SetHttpInfoState set peer`s httpinfo state
func (this *Peer) SetHttpInfoState(httpInfo bool) {
	if httpInfo {
		this.cap[common.HTTP_INFO_FLAG] = 0x01
	} else {
		this.cap[common.HTTP_INFO_FLAG] = 0x00
	}
}

//GetHttpInfoState return peer`s httpinfo state
func (this *Peer) GetHttpInfoState() bool {
	return this.cap[common.HTTP_INFO_FLAG] == 1
}

//GetHttpInfoPort return peer`s httpinfo port
func (this *Peer) GetHttpInfoPort() uint16 {
	return this.base.GetHttpInfoPort()
}

//SetHttpInfoPort set peer`s httpinfo port
func (this *Peer) SetHttpInfoPort(port uint16) {
	this.base.SetHttpInfoPort(port)
}

//UpdateInfo update peer`s information
func (this *Peer) UpdateInfo(t time.Time, version uint32, services uint64,
	syncPort uint16, consPort uint16, nonce uint64, relay uint8, height uint64, softVer string) {

	this.SyncLink.UpdateRXTime(t)
	this.base.SetID(nonce)
	this.base.SetVersion(version)
	this.base.SetServices(services)
	this.base.SetSyncPort(syncPort)
	this.base.SetConsPort(consPort)
	this.base.SetSoftVersion(softVer)
	this.SyncLink.SetPort(syncPort)
	this.ConsLink.SetPort(consPort)
	if relay == 0 {
		this.base.SetRelay(false)
	} else {
		this.base.SetRelay(true)
	}
	this.SetHeight(uint64(height))
}
