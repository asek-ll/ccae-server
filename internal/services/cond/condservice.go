package cond

import (
	"context"
	"time"

	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type CondService struct {
	clients *wsmethods.ClientsManager
}

func NewCondService(clients *wsmethods.ClientsManager) *CondService {
	return &CondService{
		clients: clients,
	}
}

func (s *CondService) getClientForCheck(check string) *wsmethods.GenericClient {
	clients := s.clients.GetClients()
	for _, client := range clients {
		gclient := client.GetGenericClient()
		if gclient.Role == check {
			return gclient
		}
	}
	return nil
}

func (s *CondService) Check(check string, params any) (bool, error) {

	client := s.getClientForCheck(check)
	if client == nil {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var res bool
	err := client.WS.SendRequestSync(ctx, "check", map[string]any{
		"cond":   check,
		"params": params,
	}, &res)
	if err != nil {
		return false, err
	}
	return res, nil
}
