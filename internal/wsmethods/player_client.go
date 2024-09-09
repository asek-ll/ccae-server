package wsmethods

import (
	"context"
	"sync"
	"time"
)

type ItemStack struct {
	Name         string `json:"name"`
	Count        int    `json:"count"`
	MaxStackSize int    `json:"maxStackSize"`
}

type PlayerClient struct {
	GenericClient

	mu sync.Mutex
}

func NewPlayerClient(base GenericClient) *PlayerClient {
	return &PlayerClient{
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
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res int
	err := s.WS.SendRequestSync(ctx, "removeItemFromPlayer", []int{slot}, &res)
	return res, err
}

func (s *PlayerClient) RemoveItems(slots []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res int
	return s.WS.SendRequestSync(ctx, "removeItemFromPlayer", slots, &res)
}

func (s *PlayerClient) AddItems(slots []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res int
	return s.WS.SendRequestSync(ctx, "addItemToPlayer", slots, &res)
}
