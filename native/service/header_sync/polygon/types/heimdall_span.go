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
package types

type HeimdallSpan struct {
	ID                uint64               `protobuf:"varint,1,opt,name=id,proto3" json:"id" yaml:"id"`
	StartBlock        uint64               `protobuf:"varint,2,opt,name=start_block,json=startBlock,proto3" json:"start_block" yaml:"start_block"`
	EndBlock          uint64               `protobuf:"varint,3,opt,name=end_block,json=endBlock,proto3" json:"end_block" yaml:"end_block"`
	ValidatorSet      HeimdallValidatorSet `protobuf:"bytes,4,opt,name=validator_set,json=validatorSet,proto3" json:"validator_set" yaml:"validator_set"`
	SelectedProducers []HeimdallValidator  `protobuf:"bytes,5,rep,name=selected_producers,json=selectedProducers,proto3" json:"selected_producers" yaml:"selected_producers"`
	BorChainId        string               `protobuf:"bytes,6,opt,name=bor_chain_id,json=borChainId,proto3" json:"bor_chain_id" yaml:"bor_chain_id"`
}

type HeimdallValidatorSet struct {
	Validators       []*HeimdallValidator `protobuf:"bytes,1,rep,name=validators,proto3" json:"validators,omitempty"`
	Proposer         *HeimdallValidator   `protobuf:"bytes,2,opt,name=proposer,proto3" json:"proposer,omitempty"`
	TotalVotingPower int64                `protobuf:"varint,3,opt,name=total_voting_power,json=totalVotingPower,proto3" json:"total_voting_power,omitempty" yaml:"total_voting_power"`
}

type ValidatorID int32

type HeimdallValidator struct {
	ID               ValidatorID `protobuf:"varint,1,opt,name=ID,proto3,enum=heimdall.types.ValidatorID" json:"ID,omitempty"`
	StartEpoch       uint64      `protobuf:"varint,2,opt,name=start_epoch,json=startEpoch,proto3" json:"start_epoch,omitempty" yaml:"start_epoch"`
	EndEpoch         uint64      `protobuf:"varint,3,opt,name=end_epoch,json=endEpoch,proto3" json:"end_epoch,omitempty" yaml:"end_epoch"`
	Nonce            uint64      `protobuf:"varint,4,opt,name=nonce,proto3" json:"nonce,omitempty"`
	VotingPower      int64       `protobuf:"varint,5,opt,name=voting_power,json=votingPower,proto3" json:"voting_power,omitempty" yaml:"voting_power"`
	PubKey           string      `protobuf:"bytes,6,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty" yaml:"pub_key"`
	Signer           string      `protobuf:"bytes,7,opt,name=signer,proto3" json:"signer,omitempty"`
	LastUpdated      string      `protobuf:"bytes,8,opt,name=last_updated,json=lastUpdated,proto3" json:"last_updated,omitempty" yaml:"last_updated"`
	Jailed           bool        `protobuf:"varint,9,opt,name=jailed,proto3" json:"jailed,omitempty"`
	ProposerPriority int64       `protobuf:"varint,10,opt,name=proposer_priority,json=proposerPriority,proto3" json:"proposer_priority,omitempty" yaml:"proposer_priority"`
}
