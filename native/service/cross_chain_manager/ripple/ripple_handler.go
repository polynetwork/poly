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
	raw, err := GetTxJsonInfo(service, params.FromChainId, params.TxHash)
	if err != nil {
		return fmt.Errorf("MultiSign, get txJsonInfo error")
	}

	// check if aleady done
	multisignInfo, err := GetMultisignInfo(service, raw)
	if err != nil {
		return fmt.Errorf("MultiSign, GetMultisignInfo error")
	}
	if multisignInfo.Status {
		return nil
	}

	// check if signature is valid
	txJson := new(types.MultisignPayment)
	err = json.Unmarshal([]byte(params.TxJson), txJson)
	if err != nil {
		return fmt.Errorf("MultiSign, unmarshal signed txjson error: %s", err)
	}
	for _, s := range txJson.Signers {
		signerAccount, err := data.NewAccountFromAddress(s.Signer.Account)
		if err != nil {
			return fmt.Errorf("MultiSign, data.NewAccountFromAddress error: %s", err)
		}
		signerPk, err := hex.DecodeString(s.Signer.SigningPubKey)
		if err != nil {
			return fmt.Errorf("MultiSign, hex.DecodeString signer pk error: %s", err)
		}
		signature, err := hex.DecodeString(s.Signer.TxnSignature)
		if err != nil {
			return fmt.Errorf("MultiSign, hex.DecodeString signature error: %s", err)
		}

		// check if valid signer
		flag := false
		for _, v := range rippleExtraInfo.Pks {
			if fmt.Sprintf("%X", v) == s.Signer.SigningPubKey {
				flag = true
				break
			}
		}
		if !flag {
			return fmt.Errorf("MultiSign, signer is not multisign account")
		}

		//check if valid signature
		err = types.CheckMultiSign(raw, *signerAccount, signerPk, signature)
		if err != nil {
			return fmt.Errorf("MultiSign, types.CheckMultiSign error: %s", err)
		}
		signer := &Signer{
			Account:       signerAccount.Bytes(),
			TxnSignature:  signature,
			SigningPubKey: signerPk,
		}
		sink := common.NewZeroCopySink(nil)
		signer.Serialization(sink)
		multisignInfo.SigMap[hex.EncodeToString(sink.Bytes())] = true
	}

	if uint64(len(multisignInfo.SigMap)) >= rippleExtraInfo.Quorum {
		payment, err := types.DeserializeRawMultiSignTx(raw)
		if err != nil {
			return fmt.Errorf("MultiSign, types.DeserializeRawMultiSignTx error")
		}
		for s := range multisignInfo.SigMap {
			signerBytes, err := hex.DecodeString(s)
			if err != nil {
				return fmt.Errorf("MultiSign, hex.DecodeString signer bytes error")
			}
			signer := new(Signer)
			err = signer.Deserialization(common.NewZeroCopySource(signerBytes))
			if err != nil {
				return fmt.Errorf("MultiSign, deserialization signer bytes error")
			}
			sig := data.Signer{}
			sig.Signer.SigningPubKey = new(data.PublicKey)
			sig.Signer.TxnSignature = new(data.VariableLength)
			*sig.Signer.TxnSignature = signer.TxnSignature
			copy(sig.Signer.SigningPubKey[:], signer.SigningPubKey)
			acc := data.Account{}
			copy(acc[:], signer.Account)
			sig.Signer.Account = acc
			payment.Signers = append(payment.Signers, sig)
		}

		finalPayment, err := json.Marshal(payment)
		if err != nil {
			return fmt.Errorf("MultiSign, json.Marshal final payment error: %s", err)
		}
		service.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.CrossChainManagerContractAddress,
				States: []interface{}{"multisignedTxJson", params.FromChainId, params.ToChainId,
					hex.EncodeToString(params.TxHash), string(finalPayment), payment.Sequence},
			})
		multisignInfo.Status = true
	}
	PutMultisignInfo(service, raw, multisignInfo)
	return nil
}

func (this *RippleHandler) MakeTransaction(service *native.NativeService, param *crosscommon.MakeTxParam,
	fromChainID uint64) error {
	source := common.NewZeroCopySource(param.Args)
	toAddrBytes, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("ripple MakeTransaction, deserialize toAddr error")
	}
	amount_temp, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("ripple MakeTransaction, deserialize amount error")
	}
	amount, err := data.NewAmount(new(big.Int).SetUint64(amount_temp).String())
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, data.NewAmount error: %s", err)
	}
	assetAddress := param.ToContractAddress

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
	if baseFee.View == 0 {
		return fmt.Errorf("ripple MakeTransaction, base fee is not initialized")
	}

	//fee = baseFee * signerNum
	fee_temp := new(big.Int).Mul(baseFee.Fee, new(big.Int).SetUint64(rippleExtraInfo.SignerNum))
	fee, err := data.NewValue(ToStringByPrecise(fee_temp, 6), true)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, data.NewValue fee error: %s", err)
	}
	feeAmount, err := data.NewAmount(fee_temp.String())
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, data.NewAmount fee error: %s", err)
	}
	amountD, err := amount.Subtract(feeAmount)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, amount.Subtract fee error: %s", err)
	}

	from := new(data.Account)
	to := new(data.Account)
	copy(from[:], assetAddress)
	copy(to[:], toAddrBytes)

	payment := types.GeneratePayment(*from, *to, *amountD, *fee, uint32(rippleExtraInfo.Sequence))
	_, raw, err := data.Raw(payment)
	if err != nil {
		return fmt.Errorf("ripple MakeTransaction, data.Raw error: %s", err)
	}
	service.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainManagerContractAddress,
			States: []interface{}{"rippleTxJson", fromChainID, param.ToChainID,
				hex.EncodeToString(param.TxHash), hex.EncodeToString(raw), payment.Sequence},
		})

	//sequence + 1
	rippleExtraInfo.Sequence = rippleExtraInfo.Sequence + 1
	side_chain_manager.PutRippleExtraInfo(service, rippleExtraInfo)

	//store txJson info
	PutTxJsonInfo(service, fromChainID, param.TxHash, hex.EncodeToString(raw))
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
	if fee.View == 0 {
		return fmt.Errorf("ReconstructTx, base fee is not initialized")
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
			States: []interface{}{"rippleTxJson", params.FromChainId, params.ToChainId,
				hex.EncodeToString(params.TxHash), txJsonStr, txJson.Sequence},
		})
	return nil
}
