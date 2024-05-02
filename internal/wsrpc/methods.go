package wsrpc

import (
	"encoding/json"
	"fmt"

	"github.com/asek-ll/aecc-server/internal/app"
)

type LoginParams struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

func withInnerId[T any](mapper *IdMapper, f func(id string, params T) (any, error)) RpcMethod {
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

func SetupMethods(server *JsonRpcServer, app *app.App) {

	idMapper := NewIdMapper()

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

	server.AddMethod("login", Typed(func(clientId uint, params LoginParams) (any, error) {
		fmt.Println("Login", params)
		err := app.Daos.Clients.LoginClient(params.ID, params.Role)
		if err != nil {
			return nil, err
		}
		idMapper.Add(params.ID, clientId)

		// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		// var res any
		// err = server.SendRequestSync(ctx, clientId, "eval", "return 2+2", &res)
		// if err != nil {
		// 	fmt.Println("Error", err)
		// } else {
		// 	fmt.Println("SUCCESS", res)
		// }

		return "OK", nil
	}))

	server.AddMethod("myId", withInnerId(idMapper, func(innerId string, params LoginParams) (any, error) {
		return innerId, nil
	}))
}
