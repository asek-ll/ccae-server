package wsmethods

import (
	"context"
)

type ModemClient struct {
	GenericClient
}

func NewModemClient(base GenericClient) *ModemClient {
	return &ModemClient{
		GenericClient: base,
	}
}

// callRemote(remoteName, method, ...)
// func (c *ModemClient) CallRemote(ctx context.Context, result interface{}, method string, params ...any) error {
// 	return c.WS.SendRequestSync(ctx, "callRemote", params, &result)
// }

// getNamesRemote()
func (c *ModemClient) GetNamesRemote(ctx context.Context) ([]string, error) {
	var res []string
	err := c.WS.SendRequestSync(ctx, "getNamesRemote", nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// getMethodsRemote(name)
func (c *ModemClient) GetMethodsRemote(ctx context.Context, name string) ([]string, error) {
	var res []string
	err := c.WS.SendRequestSync(ctx, "getMethodsRemote", name, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
