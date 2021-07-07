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
package polygon

type Span struct {
	ID                uint64       `protobuf:"varint,1,opt,name=id,proto3" json:"id" yaml:"id"`
	StartBlock        uint64       `protobuf:"varint,2,opt,name=start_block,json=startBlock,proto3" json:"start_block" yaml:"start_block"`
	EndBlock          uint64       `protobuf:"varint,3,opt,name=end_block,json=endBlock,proto3" json:"end_block" yaml:"end_block"`
	ValidatorSet      ValidatorSet `protobuf:"bytes,4,opt,name=validator_set,json=validatorSet,proto3" json:"validator_set" yaml:"validator_set"`
	SelectedProducers []Validator  `protobuf:"bytes,5,rep,name=selected_producers,json=selectedProducers,proto3" json:"selected_producers" yaml:"selected_producers"`
	BorChainId        string       `protobuf:"bytes,6,opt,name=bor_chain_id,json=borChainId,proto3" json:"bor_chain_id" yaml:"bor_chain_id"`
}
