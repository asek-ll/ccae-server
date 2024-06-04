package wsrpc

import (
	"context"
)

type ClientWrapper struct {
	server   *JsonRpcServer
	clientID uint
}

func NewClientWrapper(server *JsonRpcServer, clientId uint) ClientWrapper {
	return ClientWrapper{
		server:   server,
		clientID: clientId,
	}
}

func (h ClientWrapper) SendRequest(method string, params any) (uint, error) {
	return h.server.SendRequest(h.clientID, method, params)
}

func (h ClientWrapper) SendRequestSync(ctx context.Context, method string, params any, result interface{}) error {
	return h.server.SendRequestSync(ctx, h.clientID, method, params, result)
}
