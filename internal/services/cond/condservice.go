package cond

import (
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

func (s *CondService) Check(check string, params any) (bool, error) {
	client, err := wsmethods.GetClientForType[wsmethods.CondClient](s.clients)
	if err != nil {
		return false, err
	}

	return client.Check(check, params)
}
