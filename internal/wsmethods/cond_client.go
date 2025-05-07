package wsmethods

import (
	"context"
	"time"
)

type CondClient struct {
	GenericClient
}

func NewCondClient(base GenericClient) *CondClient {
	return &CondClient{
		GenericClient: base,
	}
}

func (c *CondClient) Check(cond string, params any) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var res bool
	err := c.WS.SendRequestSync(ctx, "check", map[string]any{
		"cond":   cond,
		"params": params,
	}, &res)
	if err != nil {
		return false, err
	}
	return res, nil
}
