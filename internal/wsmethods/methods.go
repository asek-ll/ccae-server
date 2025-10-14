package wsmethods

import (
	"encoding/json"
	"fmt"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/ws"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

type LoginParams struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type LoginV2Params struct {
	ID      int    `json:"id"`
	Version string `json:"version"`
	Role    string `json:"role"`
}

type LoginV3Params struct {
	ID      int    `json:"id"`
	Label   string `json:"label"`
	Version string `json:"version"`
}

func withInnerId[T any](mapper *wsrpc.IdMapper, f func(id string, params T) (any, error)) wsrpc.RpcMethod {
	return func(client *ws.Client, params []byte) (any, error) {
		id, e := mapper.ToInner(client.ID)
		if !e {
			return nil, fmt.Errorf("Can't find mapping for outer id: %d", client.ID)
		}

		var ps T
		err := json.Unmarshal(params, &ps)
		if err != nil {
			return nil, err
		}

		return f(id, ps)
	}
}

func SetupMethods(server *wsrpc.JsonRpcServer, clientsDao *dao.ClientsDao) {

	idMapper := wsrpc.NewIdMapper()

	server.AddMethod("ping", func(client *ws.Client, params []byte) (any, error) {
		return "pong", nil
	})

	server.AddMethod("myId", withInnerId(idMapper, func(innerId string, params LoginParams) (any, error) {
		return innerId, nil
	}))
}
