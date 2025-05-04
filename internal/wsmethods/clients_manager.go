package wsmethods

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/asek-ll/aecc-server/internal/build"
	"github.com/asek-ll/aecc-server/internal/common"
	"github.com/asek-ll/aecc-server/internal/config"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/clientscripts"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

const CLIENT_ROLE_STORAGE = "storage"
const CLIENT_ROLE_CRAFTER = "crafter"
const CLIENT_ROLE_PROCESSING = "processing"
const CLIENT_ROLE_PLAYER = "player"
const CLIENT_ROLE_MODEM = "modem"

type Client interface {
	GetGenericClient() *GenericClient
}

type GenericClient struct {
	ID       string
	Role     string
	JoinTime time.Time
	Props    map[string]any
	WS       wsrpc.ClientWrapper
}

func (c *GenericClient) GetGenericClient() *GenericClient {
	return c
}

type ClientsManager struct {
	server         *wsrpc.JsonRpcServer
	clientsDao     *dao.ClientsDao
	clients        map[uint]Client
	clientListener ClientListener

	clientIdByType map[string]uint

	mu sync.RWMutex
}

func NewClientsManager(
	server *wsrpc.JsonRpcServer,
	clientsDao *dao.ClientsDao,
	configLoader *config.ConfigLoader,
	scriptsManager *clientscripts.ScriptsManager,
) *ClientsManager {
	clientsManager := &ClientsManager{
		server:         server,
		clientsDao:     clientsDao,
		clients:        make(map[uint]Client),
		clientListener: DumpCycleListener{},
		clientIdByType: make(map[string]uint),
	}

	server.SetDisconnectHandler(func(clientId uint) error {
		return clientsManager.RemoveClient(clientId)
	})

	scriptsManager.SetOnUpdate(func(role string, content string) error {

		log.Println("[WARN] ON UPDATE!!!", content)
		log.Println("[WARN] check client", clientsManager.GetClients())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		for _, client := range clientsManager.GetClients() {
			gc := client.GetGenericClient()
			if gc.Role == role {
				props := make(map[string]any)
				err := gc.WS.SendRequestSync(ctx, "init", content, &props)
				if err != nil {
					return err
				}
				clientsManager.mu.Lock()
				gc.Props = props
				clientsManager.mu.Unlock()
			}
		}

		return nil
	})

	server.AddMethod("login", wsrpc.Typed(func(clientId uint, params LoginParams) (any, error) {
		url := fmt.Sprintf("%s/static/lua/%s.lua", configLoader.Config.WebServer.Url, params.Role)

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

	server.AddMethod("login.v2", wsrpc.Typed(func(clientId uint, params LoginV2Params) (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		if params.Version != build.Time {

			url := fmt.Sprintf("%s/lua/v2/client/%s/", configLoader.Config.WebServer.Url, params.Role)
			props := make(map[string]any)
			err := server.SendRequestSync(ctx, clientId, "upgrade", url, &props)
			if err != nil {
				return nil, err
			}

			return "OK", nil
		}

		script, err := scriptsManager.GetScript(params.Role)
		if err != nil {
			return nil, err
		}

		props := make(map[string]any)
		err = server.SendRequestSync(ctx, clientId, "init", script.Content, &props)
		// err = server.SendRequestSync(ctx, clientId, "init", map[string]any{
		// 	"content": script.Content,
		// 	"version": script.Version,
		// }, &props)
		log.Printf("[WARN] Client try register!!!")
		if err != nil {
			log.Printf("[ERROR] Can't init client: %v", err)
			return nil, err
		}

		err = clientsManager.RegisterClient(clientId, fmt.Sprintf("%d", params.ID), params.Role, props)
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

	var err error
	switch role {
	case CLIENT_ROLE_STORAGE:
		client, err = NewStorageClient(genericClient)
	case CLIENT_ROLE_CRAFTER:
		client, err = NewCrafterClient(genericClient)
	case CLIENT_ROLE_PROCESSING:
		client, err = NewCrafterClient(genericClient)
	case CLIENT_ROLE_PLAYER:
		client = NewPlayerClient(genericClient)
	case CLIENT_ROLE_MODEM:
		client = NewModemClient(genericClient)
	default:
		client = &genericClient
	}
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.clients[webscoketClientId] = client
	c.clientIdByType[role] = webscoketClientId
	c.mu.Unlock()

	c.clientListener.HandleClientConnected(client)
	return nil
}

func (c *ClientsManager) RemoveClient(id uint) error {
	c.mu.RLock()
	c.clientListener.HandleClientDisconnected(c.clients[id])
	c.mu.RUnlock()

	c.mu.Lock()
	delete(c.clients, id)
	c.mu.Unlock()
	return nil
}

func (c *ClientsManager) SetClientListener(listener ClientListener) {
	c.clientListener = listener
}

func GetClientForType[T interface{}](c *ClientsManager) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, client := range c.clients {
		convertedClient, ok := client.(T)
		if ok {
			return convertedClient, nil
		}
	}
	var empty T
	return empty, errors.New("Client not found")
}

func CallWithClientForType[T any, V any](c *ClientsManager, fn func(client T) (V, error)) (V, error) {
	client, err := GetClientForType[T](c)
	if err != nil {
		var empty V
		return empty, err
	}
	return fn(client)
}

func (c *ClientsManager) GetClients() []Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return common.MapValues(c.clients)
}
