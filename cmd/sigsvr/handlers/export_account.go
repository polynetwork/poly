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
	"encoding/json"
	"fmt"
	"github.com/polynetwork/poly/account"
	clisvrcom "github.com/polynetwork/poly/cmd/sigsvr/common"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/log"
	"io/ioutil"
	"strings"
	"time"
)

type ExportAccountReq struct {
	WalletPath string `json:"wallet_path"`
}

type ExportAccountResp struct {
	WalletFile    string `json:"wallet_file"`
	AccountNumber int    `json:"account_num"`
}

func ExportAccount(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	expReq := &ExportAccountReq{}
	err := json.Unmarshal(req.Params, expReq)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		log.Infof("ExportAccount Qid:%s json.Unmarshal ExportAccountReq error:%s", req.Qid, err)
		return
	}
	walletPath := expReq.WalletPath
	if walletPath != "" {
		if !common.FileExisted(walletPath) {
			resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
			resp.ErrorInfo = "wallet path doesn't exist"
			return
		}
	} else {
		walletPath = "./"
	}

	walletStore := clisvrcom.DefWalletStore
	walletData := &account.WalletData{
		Name:     walletStore.WalletName,
		Version:  walletStore.WalletVersion,
		Scrypt:   walletStore.WalletScrypt,
		Accounts: make([]*account.AccountData, 0),
		Extra:    walletStore.WalletExtra,
	}

	accountCount := walletStore.GetNextAccountIndex()
	for i := uint32(0); i < accountCount; i++ {
		accData, err := walletStore.GetAccountDataByIndex(i)
		if err != nil {
			log.Errorf("ExportAccount Qid:%s GetAccountDataByIndex:%d error:%s\n", req.Qid, i, err)
			continue
		}
		if accData == nil {
			continue
		}
		walletData.Accounts = append(walletData.Accounts, accData)
	}

	data, err := json.Marshal(walletData)
	if err != nil {
		log.Errorf("ExportAccount Qid:%s json.Marshal WalletData error:%s\n", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}

	walletFile := fmt.Sprintf("%s/wallet_%s.dat", strings.TrimRight(walletPath, "/"), time.Now().Format("2006_01_02_15_04_05"))
	err = ioutil.WriteFile(walletFile, data, 0666)
	if err != nil {
		log.Errorf("ExportAccount Qid:%s write file:%s error:%s", req.Qid, walletFile, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}

	resp.Result = &ExportAccountResp{
		WalletFile:    walletFile,
		AccountNumber: len(walletData.Accounts),
	}
	log.Infof("ExportAccount Qid:%s success wallet file:%s", req.Qid, walletFile)
}
