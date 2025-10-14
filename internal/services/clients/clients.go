package clients

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/asek-ll/aecc-server/internal/dao"
)

type ClientsService struct {
	clientsDao *dao.ClientsDao
}

func NewClientsService(clientsDao *dao.ClientsDao) *ClientsService {
	return &ClientsService{
		clientsDao: clientsDao,
	}
}

func (s *ClientsService) GetClientBySecret(secret string) (*dao.Client, error) {
	hash := sha256.Sum256([]byte(secret))
	clientID := hex.EncodeToString(hash[:])

	client, err := s.clientsDao.GetClientByID(clientID)
	if err != nil {
		return nil, fmt.Errorf("GetClientByID: %w", err)
	}

	if client != nil {
		return client, nil
	}

	client = &dao.Client{
		ID: clientID,
	}

	err = s.clientsDao.CreateClient(client)
	if err != nil {
		return nil, fmt.Errorf("CreateClient: %w", err)
	}

	return client, nil
}
