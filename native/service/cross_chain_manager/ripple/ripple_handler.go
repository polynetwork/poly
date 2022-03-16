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
	"crypto/sha256"
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

	// get quorum
	op, err := side_chain_manager.GetAssetMapIndex(service, params.ChainId, params.AssetAddress)
	if err != nil {
		return fmt.Errorf("MultiSign, get asset map index error: %s", err)
	}
	assetMap, err := side_chain_manager.GetAssetMap(service, op)
	if err != nil {
		return fmt.Errorf("MultiSign, get asset map error: %s", err)
	}
	assetInfo := assetMap.AssetMap[params.ChainId]
	rippleExtraInfo := new(side_chain_manager.RippleExtraInfo)
	err = rippleExtraInfo.Deserialization(common.NewZeroCopySource(assetInfo.ExtraInfo))
	if err != nil {
		return fmt.Errorf("MultiSign, deserialize rippleExtraInfo error")
	}

	// get raw txJsonInfo
	txJsonInfo, err := GetTxJsonInfo(service, params.Id)
	if err != nil {
		return fmt.Errorf("MultiSign, deserialize txJsonInfo error")
	}

	multisignInfo, err := GetMultisignInfo(service, params.Id)
	if err != nil {
		return fmt.Errorf("MultiSign, GetMultisignInfo error")
	}
	if multisignInfo.Status {
		return nil
	}
	multisignInfo.SigMap[params.TxJson] = true
	//TODO: what if fake sign is more than quorum
	if uint32(len(multisignInfo.SigMap)) >= rippleExtraInfo.Quorum {
		txJson := &types.MultisignPayment{
			Signers: make([]*types.Signer, rippleExtraInfo.Quorum),
		}
		err := json.Unmarshal([]byte(txJsonInfo.TxJson), txJson)
		if err != nil {
			return fmt.Errorf("MultiSign, unmarshal raw txjson error: %s", err)
		}
		for sig := range multisignInfo.SigMap {
			txJsonTemp := &types.MultisignPayment{
				Signers: make([]*types.Signer, rippleExtraInfo.Quorum),
			}
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
				States:          []interface{}{"multisignedTxJson", txJsonInfo.FromChainId, params.ChainId, string(txJsonFinal)},
			})
		multisignInfo.Status = true
	}
	PutMultisignInfo(service, params.Id, multisignInfo)

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

	op, err := side_chain_manager.GetAssetMapIndex(service, fromChainID, param.FromContractAddress)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, get asset map index error: %s", err)
	}
	assetMap, err := side_chain_manager.GetAssetMap(service, op)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, get asset map error: %s", err)
	}

	assetInfo := assetMap.AssetMap[param.ToChainID]
	rippleExtraInfo := new(side_chain_manager.RippleExtraInfo)
	err = rippleExtraInfo.Deserialization(common.NewZeroCopySource(assetInfo.ExtraInfo))
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, deserialize rippleExtraInfo error")
	}
	//fee = baseFee * quorum * 2
	fee := new(big.Int).Mul(rippleExtraInfo.Fee, new(big.Int).SetUint64(uint64(rippleExtraInfo.SignerNum)))

	from := new(data.Account)
	to := new(data.Account)
	copy(from[:], assetInfo.AssetAddress)
	copy(to[:], toAddrBytes)
	txJson, err := types.GenerateMultisignPaymentTxJson(from.String(), to.String(), amount, fee.String(), rippleExtraInfo.Sequence)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, GenerateMultisignPaymentTxJson error: %s", err)
	}
	id := sha256.Sum256([]byte(txJson))
	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States: []interface{}{"rippleTxJson", fromChainID, hex.EncodeToString(param.FromContractAddress),
				param.ToChainID, hex.EncodeToString(id[:]), txJson},
		})

	//sequence + 1
	rippleExtraInfo.Sequence = rippleExtraInfo.Sequence + 1
	sink := common.NewZeroCopySink(nil)
	rippleExtraInfo.Serialization(sink)
	assetMap.AssetMap[param.ToChainID].ExtraInfo = sink.Bytes()
	side_chain_manager.PutAssetMap(service, assetMap)

	//store txJson info
	txJsonInfo := &TxJsonInfo{
		TxJson:      txJson,
		FromChainId: fromChainID,
	}
	PutTxJsonInfo(service, id[:], txJsonInfo)
	return nil
}
