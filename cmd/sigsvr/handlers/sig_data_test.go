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
package handlers

import (
	"encoding/hex"
	"encoding/json"
	clisvrcom "github.com/polynetwork/poly/cmd/sigsvr/common"
	"testing"
)

func TestSigData(t *testing.T) {
	defAcc, err := testWallet.GetDefaultAccount(pwd)
	if err != nil {
		t.Errorf("GetDefaultAccount error:%s", err)
		return
	}

	rawData := []byte("HelloWorld")
	rawReq := &SigDataReq{
		RawData: hex.EncodeToString(rawData),
	}
	data, err := json.Marshal(rawReq)
	if err != nil {
		t.Errorf("json.Marshal SigDataReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:     "t",
		Method:  "sigdata",
		Params:  data,
		Account: defAcc.Address.ToBase58(),
		Pwd:     string(pwd),
	}
	resp := &clisvrcom.CliRpcResponse{}
	SigData(req, resp)
	if resp.ErrorCode != 0 {
		t.Errorf("SigData failed. ErrorCode:%d", resp.ErrorCode)
		return
	}
}
