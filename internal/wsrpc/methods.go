package wsrpc

import (
	"fmt"

	"github.com/asek-ll/aecc-server/internal/app"
)

type LoginParams struct {
	ID   string `json:"id"`
	Role string `json:"role"`
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
		return "OK", nil
	}))
}
