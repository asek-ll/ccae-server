package wsmethods

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

const CLIENT_ROLE_STORAGE = "storage"
const CLIENT_ROLE_CRAFTER = "crafter"

type Client interface {
	GetId() string
}

type GenericClient struct {
	ID string
	WS wsrpc.ClientWrapper
}

func (c *GenericClient) GetId() string {
	return c.ID
}

type ClientsManager struct {
	server     *wsrpc.JsonRpcServer
	clientsDao *dao.ClientsDao
	clients    map[uint]Client
}

func NewClientsManager(server *wsrpc.JsonRpcServer, clientsDao *dao.ClientsDao) *ClientsManager {
	clientsManager := &ClientsManager{
		server:     server,
		clientsDao: clientsDao,
		clients:    make(map[uint]Client),
	}

	// SetupMethods(server, clientsDao)

	server.SetDisconnectHandler(func(clientId uint) error {
		return clientsManager.RemoveClient(clientId)
	})

	server.AddMethod("login", wsrpc.Typed(func(clientId uint, params LoginParams) (any, error) {
		url := fmt.Sprintf("http://localhost:3001/static/lua/%s.lua", params.Role)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		props := make(map[string]any)
		err := server.SendRequestSync(ctx, clientId, "init", url, &props)
		if err != nil {
			log.Printf("[ERROR] Can't init client: %v", err)
			return nil, err
		}

		err = clientsManager.RegisterClient(clientId, params.ID, params.Role, props)
		if err != nil {
			return nil, err
		}

		return "OK", nil
	}))

	return clientsManager
}

func (c *ClientsManager) RegisterClient(webscoketClientId uint, id string, role string, props map[string]any) error {

	ws := wsrpc.NewClientWrapper(c.server, webscoketClientId)

	var client Client

	switch role {
	case CLIENT_ROLE_STORAGE:
		client = NewStorageClient(id, ws)
	case CLIENT_ROLE_CRAFTER:
		var err error
		client, err = NewCrafterClient(id, ws, props)
		if err != nil {
			return err
		}
	default:
		client = &GenericClient{ID: id, WS: ws}
	}

	c.clients[webscoketClientId] = client
	return nil
}

func (c *ClientsManager) RemoveClient(id uint) error {
	delete(c.clients, id)
	return nil
}

func (c *ClientsManager) GetStorage() (*StorageClient, error) {
	for _, client := range c.clients {
		storageClient, ok := client.(*StorageClient)
		if ok {
			return storageClient, nil
		}
	}
	return nil, nil
}

func (c *ClientsManager) GetCrafter() (*CrafterClient, error) {
	for _, client := range c.clients {
		convertedClient, ok := client.(*CrafterClient)
		if ok {
			return convertedClient, nil
		}
	}
	return nil, nil
}
