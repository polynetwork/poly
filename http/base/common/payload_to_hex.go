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
package common

import (
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/payload"
	"github.com/polynetwork/poly/core/types"
)

type PayloadInfo interface{}

//implement PayloadInfo define BookKeepingInfo
type BookKeepingInfo struct {
	Nonce uint64
}

type InvokeCodeInfo struct {
	Code string
}
type DeployCodeInfo struct {
	Code        string
	NeedStorage bool
	Name        string
	CodeVersion string
	Author      string
	Email       string
	Description string
}

type RecordInfo struct {
	RecordType string
	RecordData string
}

type BookkeeperInfo struct {
	PubKey     string
	Action     string
	Issuer     string
	Controller string
}

type DataFileInfo struct {
	IPFSPath string
	Filename string
	Note     string
	Issuer   string
}

type PrivacyPayloadInfo struct {
	PayloadType uint8
	Payload     string
	EncryptType uint8
	EncryptAttr string
}

type VoteInfo struct {
	PubKeys []string
	Voter   string
}

//get tranasction payload data
func TransPayloadToHex(p types.Payload) PayloadInfo {
	switch object := p.(type) {
	case *payload.InvokeCode:
		obj := new(InvokeCodeInfo)
		obj.Code = common.ToHexString(object.Code)
		return obj
	}
	return nil
}
