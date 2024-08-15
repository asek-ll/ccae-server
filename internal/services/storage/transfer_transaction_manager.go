package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type TransferTransaction struct {
	DBTx     *sql.Tx
	stx      *dao.StoredTX
	tm       *TransferTransactionManager
	complete bool
}

func (tx *TransferTransaction) Commit() error {
	if tx.complete {
		return nil
	}
	tx.complete = true
	defer tx.tm.unlock()
	err := tx.DBTx.Commit()
	if err != nil {
		return err
	}

	return tx.tm.processSTX(tx.stx)
}

func (tx *TransferTransaction) Rollback() error {
	if tx.complete {
		return nil
	}
	tx.complete = true

	defer tx.tm.unlock()

	return tx.DBTx.Rollback()
}

type TransferTransactionManager struct {
	exportTxDao    *dao.StoredTXDao
	storageAdapter *wsmethods.StorageAdapter
	storage        *Storage
	mu             sync.Mutex
}

func (tm *TransferTransactionManager) unlock() {
	tm.mu.Unlock()
}

func (tm *TransferTransactionManager) processSTX(tx *dao.StoredTX) error {
	var data ExportTransactionData
	err := json.Unmarshal(tx.Data, &data)
	if err != nil {
		return err
	}
	err = tm.performTransfer(&data)
	if err != nil {
		return err
	}

	return tm.exportTxDao.DropTransaction(tx)
}

func (tm *TransferTransactionManager) restoreIfExists(subjects []string) error {
	txs, err := tm.exportTxDao.FindLocks(subjects)
	if err != nil {
		return err
	}
	if len(txs) == 0 {
		return nil
	}

	for _, tx := range txs {
		err := tm.processSTX(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tm *TransferTransactionManager) performTransfer(data *ExportTransactionData) error {
	for _, stack := range data.ItemStacks {
		_, err := tm.storageAdapter.MoveStack(stack.StorageName, stack.Slot, stack.TargetStorage, stack.ToSlot, stack.Amount)
		if err != nil {
			return err
		}
	}

	for _, stack := range data.FluidStacks {
		_, err := tm.storageAdapter.MoveFluid(stack.TankName, stack.TargetTankName, stack.Amount, stack.Uid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tm *TransferTransactionManager) setupTransaction(itemStore string, fluidStore string, request ExportRequest) (*TransferTransaction, error) {
	err := tm.restoreIfExists([]string{itemStore, fluidStore})
	if err != nil {
		return nil, err
	}

	err = tm.storage.ImportAll(itemStore)
	if err != nil {
		return nil, err
	}

	err = tm.storage.ImportAllFluids(fluidStore)
	if err != nil {
		return nil, err
	}

	data := &ExportTransactionData{}
	slot := 1
	for _, item := range request.RequestItems {
		_, err := tm.storage.ExportStack(item.Uid, itemStore, slot, item.Amount)
		slot += 1
		if err != nil {
			return nil, err
		}
		data.ItemStacks = append(data.ItemStacks, ExportTransactionStorageSlot{
			StorageName:   itemStore,
			Slot:          slot,
			TargetStorage: item.TargetStorage,
			ToSlot:        item.ToSlot,
			Amount:        item.Amount,
		})
	}

	for _, fluid := range request.RequestFluids {
		_, err := tm.storage.ExportFluid(fluid.Uid, fluidStore, fluid.Amount)
		if err != nil {
			return nil, err
		}
		data.FluidStacks = append(data.FluidStacks, ExportTransactionTank{
			TankName:       fluidStore,
			TargetTankName: fluid.TargetTankName,
			Uid:            fluid.Uid,
			Amount:         fluid.Amount,
		})
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	stx := &dao.StoredTX{
		Data: jsonData,
	}

	dbTx, err := tm.exportTxDao.InsertTransactionUncommitted([]string{itemStore, fluidStore}, stx)
	if err != nil {
		return nil, err
	}

	return &TransferTransaction{
		DBTx:     dbTx,
		stx:      stx,
		tm:       tm,
		complete: false,
	}, nil
}

func (tm *TransferTransactionManager) CreateExportTransaction(request ExportRequest) (*TransferTransaction, error) {
	if len(request.RequestItems) > 27 || len(request.RequestFluids) > 1 {
		return nil, fmt.Errorf("Too huge transaction")
	}

	tm.mu.Lock()
	tx, err := tm.setupTransaction("chest", "tank", request)
	if err != nil {
		tm.mu.Unlock()
		return nil, err
	}
	return tx, nil
}
