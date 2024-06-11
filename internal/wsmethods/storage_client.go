package wsmethods

import (
	"context"
	"strings"
	"time"

	"github.com/asek-ll/aecc-server/internal/common"
)

type StorageClient struct {
	GenericClient
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

type StackWithSlot struct {
	Item Stack `json:"item"`
	Slot int   `json:"slot"`
}

type Inventory struct {
	Name  string          `json:"name"`
	Items []StackWithSlot `json:"items"`
	Size  int             `json:"size"`
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

func NewStorageClient(base GenericClient) *StorageClient {
	return &StorageClient{
		GenericClient: base,
	}
}

type MoveStackParams struct {
	From   SlotRef
	To     SlotRef
	Amount int
}

func (s *StorageClient) MoveStack(fromInventory string, fromSlot int, toInventory string, toSlot int, amount int) (int, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	var detail StackDetail
	err := s.WS.SendRequestSync(ctx, "getItemDetail", slotRef, &detail)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (s *StorageClient) ExportStack(params []ExportParams) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.WS.SendRequestSync(ctx, "exportStack", params, nil)
}

func (s *StorageClient) ImportStack(params []ImportParams) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.WS.SendRequestSync(ctx, "importStack", params, nil)
}

func (s *StorageClient) GetItems() ([]Inventory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	var res []Inventory
	err := s.WS.SendRequestSync(ctx, "getItems", nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
