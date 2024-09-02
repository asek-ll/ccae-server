package wsmethods

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/asek-ll/aecc-server/internal/common"
)

type StorageClient struct {
	GenericClient

	InputStorages              []string
	ColdStoragePrefix          string
	WarmStoragePrefix          string
	SingleFluidContainerPrefix string
	TransactionStorage         string
	TransactionTanks           []string

	mu sync.Mutex
}

type ItemRef struct {
	Name string `json:"name"`
	NBT  string `json:"nbt"`
}

type SlotRef struct {
	InventoryName string `json:"inventoryName"`
	Slot          int    `json:"slot"`
}

type ExportParams struct {
	Item   ItemRef `json:"item"`
	Target SlotRef `json:"target"`
	Amount int     `json:"amount"`
}

type ImportParams struct {
	Target SlotRef `json:"target"`
	Amount int     `json:"amount"`
}

type Stack struct {
	Name     string `json:"name"`
	NBT      string `json:"nbt"`
	Count    int    `json:"count"`
	MaxCount int    `json:"maxCount"`
}

type StackDetail struct {
	Name     string `json:"name"`
	NBT      string `json:"nbt"`
	Count    int    `json:"count"`
	MaxCount int    `json:"maxCount"`
}

func (s Stack) GetUID() string {
	var nbt *string
	if s.NBT != "" {
		nbt = &s.NBT
	}
	return common.MakeUid(s.Name, nbt)
}

func (s StackDetail) GetUID() string {
	var nbt *string
	if s.NBT != "" {
		nbt = &s.NBT
	}
	return common.MakeUid(s.Name, nbt)
}

type StackWithSlot struct {
	Item Stack `json:"item"`
	Slot int   `json:"slot"`
}

type Inventory struct {
	Name  string          `json:"name"`
	Items []StackWithSlot `json:"items"`
	Size  int             `json:"size"`
}

type FluidStack struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

type FluidTank struct {
	Slot  int        `json:"slot"`
	Fluid FluidStack `json:"fluid"`
}

type FluidTanks map[int]FluidTank

type FluidContainer struct {
	Name  string      `json:"name"`
	Tanks []FluidTank `json:"tanks"`
}

func ItemRefFromUid(uid string) ItemRef {
	parts := strings.Split(uid, ":")
	if len(parts) == 3 && len(parts[2]) == 32 {
		return ItemRef{Name: parts[0] + ":" + parts[1],
			NBT: parts[2],
		}
	}
	return ItemRef{Name: uid}
}

func NewStorageClient(base GenericClient) (*StorageClient, error) {
	input, ok := base.Props["input"]
	if !ok {
		return nil, errors.New("Expected input storage name")
	}

	inputs, ok := input.([]interface{})
	if !ok {
		return nil, errors.New("Expected input storage name")
	}

	var inputNames []string
	for _, name := range inputs {
		inputName, ok := name.(string)
		if !ok {
			return nil, errors.New("Expected input storage name")
		}
		inputNames = append(inputNames, inputName)
	}

	coldStoragePrefix, ok := base.Props["cold_storage_prefix"].(string)
	if !ok {
		return nil, errors.New("Expected cold storage name")
	}
	warmStoragePrefix, ok := base.Props["warm_storage_prefix"].(string)
	if !ok {
		return nil, errors.New("Expected warm storage name")
	}

	singleFluidContainerPrefix, ok := base.Props["single_fluid_container_prefix"].(string)
	if !ok {
		return nil, errors.New("Expected single fluid container name")
	}

	transactionTanks, _ := base.Props["transaction_tank"].(string)
	transactionStorage, _ := base.Props["transaction_storage"].(string)

	return &StorageClient{
		GenericClient:              base,
		InputStorages:              inputNames,
		ColdStoragePrefix:          coldStoragePrefix,
		WarmStoragePrefix:          warmStoragePrefix,
		SingleFluidContainerPrefix: singleFluidContainerPrefix,
		TransactionStorage:         transactionStorage,
		TransactionTanks:           strings.Split(transactionTanks, ","),
	}, nil
}

type MoveStackParams struct {
	From   SlotRef `json:"from"`
	To     SlotRef `json:"to"`
	Amount int     `json:"amount"`
}

type MoveFluidParams struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Fluid  string `json:"fluid"`
	Amount int    `json:"amount"`
}

func (s *StorageClient) MoveStack(fromInventory string, fromSlot int, toInventory string, toSlot int, amount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	var moved int
	err := s.WS.SendRequestSync(ctx, "moveStack", MoveStackParams{
		From:   SlotRef{InventoryName: fromInventory, Slot: fromSlot},
		To:     SlotRef{InventoryName: toInventory, Slot: toSlot},
		Amount: amount,
	}, &moved)
	if err != nil {
		return 0, err
	}
	return moved, nil
}

func (s *StorageClient) GetStackDetail(slotRef SlotRef) (*StackDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	var detail StackDetail
	err := s.WS.SendRequestSync(ctx, "getItemDetail", slotRef, &detail)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (s *StorageClient) GetItems(prefixes []string) ([]Inventory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var res []Inventory
	err := s.WS.SendRequestSync(ctx, "getItems", prefixes, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *StorageClient) ListItems(inventoryName string) ([]StackWithSlot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var res []StackWithSlot
	err := s.WS.SendRequestSync(ctx, "getInventoryItems", inventoryName, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *StorageClient) GetTanks(tankName string) ([]FluidTank, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var res []FluidTank
	err := s.WS.SendRequestSync(ctx, "getTanks", tankName, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *StorageClient) GetFluidContainers(prefixes []string) ([]FluidContainer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var res []FluidContainer
	err := s.WS.SendRequestSync(ctx, "getFluidContainers", prefixes, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *StorageClient) MoveFluid(fromContainer string, toContainer string, amount int, fluidName string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	var moved int
	err := s.WS.SendRequestSync(ctx, "moveFluid", MoveFluidParams{
		From:   fromContainer,
		To:     toContainer,
		Amount: amount,
		Fluid:  fluidName,
	}, &moved)
	if err != nil {
		return 0, err
	}
	return moved, nil
}
