package wsmethods

import (
	"encoding/json"
	"fmt"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

type LoginParams struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

func withInnerId[T any](mapper *wsrpc.IdMapper, f func(id string, params T) (any, error)) wsrpc.RpcMethod {
	return func(clientId uint, params []byte) (any, error) {
		id, e := mapper.ToInner(clientId)
		if !e {
			return nil, fmt.Errorf("Can't find mapping for outer id: %d", clientId)
		}

		var ps T
		err := json.Unmarshal(params, &ps)
		if err != nil {
			return nil, err
		}

		return f(id, ps)
	}
}

func SetupMethods(server *wsrpc.JsonRpcServer, app *app.App) {

	idMapper := wsrpc.NewIdMapper()

	server.SetDisconnectHandler(func(clientId uint) error {
		id, e := idMapper.ToInner(clientId)
		if !e {
			return fmt.Errorf("Can't find mapping for outer id: %d", clientId)
		}
		err := app.Daos.Clients.LogoutClient(id)
		if err != nil {
			return err
		}
		idMapper.RemoveByOuter(clientId)
		return nil
	})

	server.AddMethod("ping", func(clientId uint, params []byte) (any, error) {
		return "pong", nil
	})

	server.AddMethod("login", wsrpc.Typed(func(clientId uint, params LoginParams) (any, error) {
		fmt.Println("Login", params)
		err := app.Daos.Clients.LoginClient(params.ID, params.Role, clientId)
		if err != nil {
			return nil, err
		}
		idMapper.Add(params.ID, clientId)

		url := fmt.Sprintf("http://localhost:3001/static/lua/%s.lua", params.Role)
		_, err = server.SendRequest(clientId, "init", url)
		if err != nil {
			fmt.Println("Error", err)
		}

		return "OK", nil
	}))

	server.AddMethod("myId", withInnerId(idMapper, func(innerId string, params LoginParams) (any, error) {
		return innerId, nil
	}))
}
