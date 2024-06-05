package wsmethods

import (
	"context"
	"time"
)

type ItemStack struct {
	Name         string `json:"name"`
	Count        int    `json:"count"`
	MaxStackSize int    `json:"maxStackSize"`
}

type PlayerClient struct {
	GenericClient
}

func NewPlayerClient(base GenericClient) *StorageClient {
	return &StorageClient{
		GenericClient: base,
	}
}

func (s *PlayerClient) GetItems() (map[int]ItemStack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res map[int]ItemStack
	err := s.WS.SendRequestSync(ctx, "getItems", nil, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *PlayerClient) RemoveItem(slot int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res int
	err := s.WS.SendRequestSync(ctx, "removeItemFromPlayer", []int{slot}, &res)
	return res, err
}

func (s *PlayerClient) RemoveItems(slots []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res int
	return s.WS.SendRequestSync(ctx, "removeItemFromPlayer", slots, &res)
}
