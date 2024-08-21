package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/asek-ll/aecc-server/internal/config"
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
	configLoader   *config.ConfigLoader
	exportTxDao    *dao.StoredTXDao
	storageAdapter *wsmethods.StorageAdapter
	storage        *Storage
	mu             sync.Mutex
}

func NewTransferTransactionManager(
	configLoader *config.ConfigLoader,
	exportTxDao *dao.StoredTXDao,
	storageAdapter *wsmethods.StorageAdapter,
	storage *Storage,
) *TransferTransactionManager {
	manager := &TransferTransactionManager{
		configLoader:   configLoader,
		exportTxDao:    exportTxDao,
		storageAdapter: storageAdapter,
		storage:        storage,
	}

	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := manager.pingTransaction()
			if err != nil {
				log.Printf("[WARN] transaction ping error %v", err)
			}
		}
	}()

	return manager
}

func (tm *TransferTransactionManager) pingTransaction() error {
	txs, err := tm.exportTxDao.FindTransactions()
	if err != nil {
		return err
	}
	for _, tx := range txs {
		err := tm.processSTX(tx)
		if err != nil {
			return err
		}
	}
	return nil
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
	storages := make(map[string]struct{})
	for _, stack := range data.ItemStacks {
		_, err := tm.storageAdapter.MoveStack(stack.StorageName, stack.Slot, stack.TargetStorage, stack.ToSlot, stack.Amount)
		if err != nil {
			return err
		}

		storages[stack.StorageName] = struct{}{}
	}
	for storage := range storages {
		items, err := tm.storageAdapter.ListItems(storage)
		if err != nil {
			return err
		}
		if len(items) > 0 {
			return fmt.Errorf("Transaction non performed (item)")
		}
	}

	tanks := make(map[string]struct{})
	for _, stack := range data.FluidStacks {
		_, err := tm.storageAdapter.MoveFluid(stack.TankName, stack.TargetTankName, stack.Amount, stack.Uid)
		if err != nil {
			return err
		}
		tanks[stack.TankName] = struct{}{}
	}
	for tank := range tanks {
		fluids, err := tm.storageAdapter.GetTanks(tank)
		if err != nil {
			return err
		}
		if len(fluids) > 0 {
			return fmt.Errorf("Transaction non performed (fluid)")
		}
	}

	return nil
}

func (tm *TransferTransactionManager) setupTransaction(itemStore string, fluidStores []string, request ExportRequest) (*TransferTransaction, error) {

	var subjects []string
	if len(request.RequestItems) > 0 {
		subjects = append(subjects, itemStore)
	}
	for i := range request.RequestFluids {
		subjects = append(subjects, fluidStores[i])
	}

	log.Printf("Restore %v", request)
	err := tm.restoreIfExists(subjects)
	if err != nil {
		return nil, err
	}

	log.Printf("Dump items before %v", request)
	if len(request.RequestItems) > 0 {
		err = tm.storage.ImportAll(itemStore)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("Dump fluid before %v", request)
	for _, fluidStore := range fluidStores {
		err = tm.storage.ImportAllFluids(fluidStore)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("FORM transaction %v", request)
	data := &ExportTransactionData{}
	slot := 0
	for _, item := range request.RequestItems {
		slot += 1
		amount, err := tm.storage.ExportStack(item.Uid, itemStore, slot, item.Amount)
		if amount != item.Amount {
			return nil, fmt.Errorf("Can't move items for stx")
		}
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

	for i, fluid := range request.RequestFluids {
		amount, err := tm.storage.ExportFluid(fluid.Uid, fluidStores[i], fluid.Amount)
		if amount != fluid.Amount {
			return nil, fmt.Errorf("Can't move fluid for stx, moved %d, but need %d", amount, fluid.Amount)
		}
		if err != nil {
			return nil, err
		}
		data.FluidStacks = append(data.FluidStacks, ExportTransactionTank{
			TankName:       fluidStores[i],
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

	dbTx, err := tm.exportTxDao.InsertTransactionUncommitted(subjects, stx)
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
	if len(request.RequestItems) == 0 && len(request.RequestFluids) == 0 {
		return nil, fmt.Errorf("Empty transaction")
	}
	if len(request.RequestItems) > 27 || len(request.RequestFluids) > 1 {
		return nil, fmt.Errorf("Too huge transaction")
	}
	client, err := tm.storageAdapter.GetClient()
	if err != nil {
		return nil, err
	}
	if len(request.RequestItems) > 0 {
		if client.TransactionStorage == "" {
			return nil, fmt.Errorf("No transaction storage specified")
		}
	}
	if len(request.RequestFluids) > len(client.TransactionTanks) {
		return nil, fmt.Errorf("No transaction storage specified")
	}
	for i := range request.RequestFluids {
		if client.TransactionTanks[i] == "" {
			return nil, fmt.Errorf("No transaction storage specified")
		}
	}

	tm.mu.Lock()
	tx, err := tm.setupTransaction(client.TransactionStorage, client.TransactionTanks, request)
	if err != nil {
		tm.mu.Unlock()
		return nil, err
	}
	return tx, nil
}
