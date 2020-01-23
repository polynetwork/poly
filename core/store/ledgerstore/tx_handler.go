/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package ledgerstore

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/core/payload"
	"github.com/ontio/multi-chain/core/store"
	scommon "github.com/ontio/multi-chain/core/store/common"
	"github.com/ontio/multi-chain/core/store/overlaydb"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/storage"
)

//HandleInvokeTransaction deal with smart contract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, cache *storage.CacheDB,
	tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) ([]common.Uint256, error) {
	invoke := tx.Payload.(*payload.InvokeCode)

	service := native.NewNativeService(cache, tx, block.Header.Timestamp, block.Header.Height,
		block.Hash(), block.Header.ChainID, invoke.Code, false)

	if _, err := service.Invoke(); err != nil {
		return nil, err
	}
	notify.Notify = append(notify.Notify, service.GetNotify()...)
	notify.State = event.CONTRACT_STATE_SUCCESS
	service.GetCacheDB().Commit()
	return service.GetCrossHashes(), nil
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notify *event.ExecuteNotify) error {
	if !config.DefConfig.Common.EnableEventLog {
		return nil
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notify); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notify)
	return nil
}
