package wsmethods

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

const CLIENT_ROLE_STORAGE = "storage"
const CLIENT_ROLE_CRAFTER = "crafter"

type Client interface {
	GetID() string
	GetRole() string
	GetJoinTime() time.Time
	GetProps() map[string]any
}

type GenericClient struct {
	ID       string
	Role     string
	JoinTime time.Time
	Props    map[string]any
	WS       wsrpc.ClientWrapper
}

func (c *GenericClient) GetID() string {
	return c.ID
}

func (c *GenericClient) GetRole() string {
	return c.Role
}

func (c *GenericClient) GetJoinTime() time.Time {
	return c.JoinTime
}

func (c *GenericClient) GetProps() map[string]any {
	return c.Props
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

	genericClient := GenericClient{
		ID:       id,
		Role:     role,
		JoinTime: time.Now(),
		Props:    props,
		WS:       ws,
	}
	var client Client

	switch role {
	case CLIENT_ROLE_STORAGE:
		client = NewStorageClient(genericClient)
	case CLIENT_ROLE_CRAFTER:
		var err error
		client, err = NewCrafterClient(genericClient)
		if err != nil {
			return err
		}
	default:
		client = &genericClient
	}

	c.clients[webscoketClientId] = client
	return nil
}

func (c *ClientsManager) RemoveClient(id uint) error {
	delete(c.clients, id)
	return nil
}

func GetClientForType[T interface{}](c *ClientsManager) (T, error) {
	var empty T
	log.Printf("[INFO] empty is %v", empty)
	for _, client := range c.clients {
		log.Printf("[INFO] check client %v", client)
		convertedClient, ok := client.(T)
		if ok {
			return convertedClient, nil
		}
	}
	return empty, errors.New("Client not found")
}

func (c *ClientsManager) GetClients() []Client {
	return common.MapValues(c.clients)
}
