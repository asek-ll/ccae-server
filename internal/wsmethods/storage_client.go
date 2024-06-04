package wsmethods

import (
	"context"
	"strings"
	"time"

	"github.com/asek-ll/aecc-server/internal/wsrpc"
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

func ItemRefFromUid(uid string) ItemRef {
	parts := strings.Split(uid, ":")
	if len(parts) == 3 && len(parts[2]) == 32 {
		return ItemRef{Name: parts[0] + ":" + parts[1],
			NBT: parts[2],
		}
	}
	return ItemRef{Name: uid}
}

func NewStorageClient(ID string, WS wsrpc.ClientWrapper) *StorageClient {
	return &StorageClient{
		GenericClient: GenericClient{
			ID: ID,
			WS: WS,
		},
	}
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
