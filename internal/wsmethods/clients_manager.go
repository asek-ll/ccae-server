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
	"github.com/asek-ll/aecc-server/internal/ws"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

const CLIENT_ROLE_STORAGE = "storage"
const CLIENT_ROLE_CRAFTER = "crafter"
const CLIENT_ROLE_PROCESSING = "processing"
const CLIENT_ROLE_PLAYER = "player"
const CLIENT_ROLE_MODEM = "modem"
const CLIENT_ROLE_COND = "cond"

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
	configLoader   *config.ConfigLoader

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
		configLoader:   configLoader,
	}

	server.SetDisconnectHandler(func(client *ws.Client) error {
		return clientsManager.RemoveClient(client.ID)
	})

	scriptsManager.SetOnUpdate(func(script *dao.ClientsScript) error {
		return clientsManager.OnUpdateScript(script)
	})

	server.AddMethod("login", wsrpc.Typed(func(wsClient *ws.Client, params LoginV3Params) (any, error) {

		client, err := clientsDao.GetClientByID(wsClient.ExtID)
		if err != nil {
			return nil, err
		}
		if client == nil || !client.Authorized {
			return nil, fmt.Errorf("Unathorized")
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		script, err := scriptsManager.GetScript("bootstrap")
		if err != nil {
			return nil, err
		}

		targetVersion := build.Time
		if script != nil {
			targetVersion = fmt.Sprintf("%d", script.Version)
		}

		if params.Version != targetVersion {
			url := fmt.Sprintf("%s/lua/v3/client/", configLoader.Config.WebServer.Url)
			var props any
			err := server.SendRequestSync(ctx, wsClient.ID, "upgrade", url, &props)
			if err != nil {
				return nil, err
			}

			return "OK", nil
		}

		err = clientsDao.LoginClient(client, wsClient.ID)
		if err != nil {
			return nil, err
		}

		if client.Role != "" {
			script, err := scriptsManager.GetScript(client.Role)
			if err != nil {
				return nil, err
			}

			if script != nil {
				var props any
				contentUrl := fmt.Sprintf("%s/clients-scripts/%s/content/", configLoader.Config.WebServer.Url, client.Role)
				err = server.SendRequestSync(ctx, wsClient.ID, "init", map[string]any{
					"contentUrl": contentUrl,
					"version":    script.Version,
				}, &props)
				log.Printf("[WARN] Client try register!!!")
				if err != nil {
					log.Printf("[ERROR] Can't init client: %v", err)
					return nil, err
				}
			}
		}

		err = clientsManager.RegisterClient(wsClient.ID, fmt.Sprintf("%d", params.ID), client.Role, nil)
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
	case CLIENT_ROLE_COND:
		client = NewCondClient(genericClient)
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

	c.clientsDao.LogoutClient(id)

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

func (c *ClientsManager) OnUpdateScript(script *dao.ClientsScript) error {
	log.Println("[WARN] ON UPDATE!!!", script.Role)
	log.Println("[WARN] check client", c.GetClients())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	url := fmt.Sprintf("%s/clients-scripts/%s/content/", c.configLoader.Config.WebServer.Url, script.Role)
	clients, err := c.clientsDao.GetActiveClientsByRole(script.Role)
	if err != nil {
		return err
	}

	for _, client := range clients {
		if client.WSClientID == nil {
			continue
		}

		wsClientWrapper, ok := c.clients[*client.WSClientID]
		if !ok {
			continue
		}
		gc := wsClientWrapper.GetGenericClient()

		props := make(map[string]any)
		err := gc.WS.SendRequestSync(ctx, "init", map[string]any{
			"version":    script.Version,
			"contentUrl": url,
		}, &props)
		if err != nil {
			return err
		}
		c.mu.Lock()
		gc.Props = props
		c.mu.Unlock()
	}

	return nil
}
