/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package ripple

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/event"
	crosscommon "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/poly/native/service/governance/side_chain_manager"
	"github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/ripple-sdk/types"
	"github.com/rubblelabs/ripple/data"
)

type RippleHandler struct {
}

func NewRippleHandler() *RippleHandler {
	return &RippleHandler{}
}

func (this *RippleHandler) MultiSign(service *native.NativeService) error {
	params := new(MultiSignParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return fmt.Errorf("MultiSign, contract params deserialize error: %v", err)
	}

	// get rippleExtraInfo
	rippleExtraInfo, err := side_chain_manager.GetRippleExtraInfo(service, params.ToChainId, params.AssetAddress)
	if err != nil {
		return fmt.Errorf("MultiSign, side_chain_manager.GetRippleExtraInfo error")
	}

	// get raw txJsonInfo
	txJsonInfo, err := GetTxJsonInfo(service, params.FromChainId, params.TxHash)
	if err != nil {
		return fmt.Errorf("MultiSign, deserialize txJsonInfo error")
	}

	multisignInfo, err := GetMultisignInfo(service, txJsonInfo)
	if err != nil {
		return fmt.Errorf("MultiSign, GetMultisignInfo error")
	}
	if multisignInfo.Status {
		return nil
	}
	multisignInfo.SigMap[params.TxJson] = true
	//TODO: what if fake sign is more than quorum
	if uint64(len(multisignInfo.SigMap)) >= rippleExtraInfo.Quorum {
		txJson := &types.MultisignPayment{
			Signers: make([]*types.Signer, 0),
		}
		err := json.Unmarshal([]byte(txJsonInfo), txJson)
		if err != nil {
			return fmt.Errorf("MultiSign, unmarshal raw txjson error: %s", err)
		}
		for sig := range multisignInfo.SigMap {
			txJsonTemp := new(types.MultisignPayment)
			err := json.Unmarshal([]byte(sig), txJsonTemp)
			if err != nil {
				return fmt.Errorf("MultiSign, unmarshal signed txjson error: %s", err)
			}
			txJson.Signers = append(txJson.Signers, txJsonTemp.Signers...)
		}
		txJsonFinal, err := json.Marshal(txJson)
		if err != nil {
			return fmt.Errorf("MultiSign, json.Marshal final txJson string error: %s", err)
		}
		service.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.CrossChainManagerContractAddress,
				States: []interface{}{"multisignedTxJson", params.FromChainId, params.ToChainId,
					hex.EncodeToString(params.TxHash), string(txJsonFinal)},
			})
		multisignInfo.Status = true
	}
	PutMultisignInfo(service, txJsonInfo, multisignInfo)

	return nil
}

func (this *RippleHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam,
	fromChainID uint64) error {
	source := common.NewZeroCopySource(param.Args)
	toAddrBytes, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("ripple MakeTransaction, deserialize toAddr error")
	}
	amount, eof := source.NextString()
	if eof {
		return fmt.Errorf("ripple MakeTransaction, deserialize amount error")
	}

	//get asset map
	op, err := side_chain_manager.GetAssetMapIndex(service, fromChainID, param.FromContractAddress)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, get asset map index error: %s", err)
	}
	assetMap, err := side_chain_manager.GetAssetMap(service, op)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, get asset map error: %s", err)
	}
	assetAddress := assetMap.AssetMap[param.ToChainID]

	// get rippleExtraInfo
	rippleExtraInfo, err := side_chain_manager.GetRippleExtraInfo(service, param.ToChainID, assetAddress)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, side_chain_manager.GetRippleExtraInfo error")
	}

	//get fee
	baseFee, err := side_chain_manager.GetFee(service, param.ToChainID)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, side_chain_manager.GetFee error: %v", err)
	}

	//fee = baseFee * signerNum
	fee := new(big.Int).Mul(baseFee.Fee, new(big.Int).SetUint64(rippleExtraInfo.SignerNum))

	from := new(data.Account)
	to := new(data.Account)
	copy(from[:], assetAddress)
	copy(to[:], toAddrBytes)

	//add memos
	memo := types.Memo{}
	memo.Memo.MemoType = "706f6c7968617368" // == "polyhash"
	polyHash := service.GetTx().Hash()
	memo.Memo.MemoData = polyHash.ToHexString()
	memos := []types.Memo{memo}
	txJson, err := types.GenerateMultisignPaymentTxJson(from.String(), to.String(), amount, fee.String(),
		uint32(rippleExtraInfo.Sequence), memos)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, GenerateMultisignPaymentTxJson error: %s", err)
	}
	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{"rippleTxJson", fromChainID, param.ToChainID, hex.EncodeToString(param.TxHash), txJson},
		})

	//sequence + 1
	rippleExtraInfo.Sequence = rippleExtraInfo.Sequence + 1
	side_chain_manager.PutRippleExtraInfo(service, rippleExtraInfo)

	//store txJson info
	PutTxJsonInfo(service, fromChainID, param.TxHash, txJson)
	return nil
}

func (this *RippleHandler) ReconstructTx(service *native.NativeService) error {
	params := new(ReconstructTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(service.GetInput())); err != nil {
		return fmt.Errorf("ReconstructTx, contract params deserialize error: %v", err)
	}

	//get tx json info
	txJsonInfo, err := GetTxJsonInfo(service, params.FromChainId, params.TxHash)
	if err != nil {
		return fmt.Errorf("ReconstructTx, GetTxJsonInfo error: %v", err)
	}

	//get fee
	fee, err := side_chain_manager.GetFee(service, params.ToChainId)
	if err != nil {
		return fmt.Errorf("ReconstructTx, side_chain_manager.GetFee error: %v", err)
	}

	//get ripple extra info
	rippleExtraInfo, err := side_chain_manager.GetRippleExtraInfo(service, params.ToChainId, params.AssetAddress)
	if err != nil {
		return fmt.Errorf("ReconstructTx, side_chain_manager.GetRippleExtraInfo error: %v", err)
	}

	txJson := new(types.MultisignPayment)
	err = json.Unmarshal([]byte(txJsonInfo), txJson)
	if err != nil {
		return fmt.Errorf("ReconstructTx, json.Unmarshal tx json info error: %v", err)
	}
	txJson.Fee = new(big.Int).Mul(fee.Fee, new(big.Int).SetUint64(rippleExtraInfo.SignerNum)).String()
	txJsonStr, err := json.Marshal(txJson)
	if err != nil {
		return fmt.Errorf("ReconstructTx, json.Marshal tx json error: %v", err)
	}
	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States:          []interface{}{"rippleTxJson", params.FromChainId, params.ToChainId, hex.EncodeToString(params.TxHash), txJsonStr},
		})
	return nil
}
