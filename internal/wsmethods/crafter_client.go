package wsmethods

import (
	"context"
	"fmt"
	"time"
)

type CrafterClient struct {
	GenericClient
	bufferName string
}

func NewCrafterClient(base GenericClient) (*CrafterClient, error) {
	bufferName, ok := base.Props["buffer_name"]
	if !ok {
		return nil, fmt.Errorf("invalid buffer_name: %v", base.Props["buffer_name"])
	}
	return &CrafterClient{
		GenericClient: base,
		bufferName:    bufferName,
	}, nil
}

func (c *CrafterClient) BufferName() string {
	return c.bufferName
}

func (c *CrafterClient) Craft() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res bool
	err := c.WS.SendRequestSync(ctx, "craft", nil, &res)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *CrafterClient) DumpOut() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res bool
	err := c.WS.SendRequestSync(ctx, "dumpOut", nil, &res)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *CrafterClient) Restore() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res bool
	err := c.WS.SendRequestSync(ctx, "restore", nil, &res)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *CrafterClient) ProcessResults(inventory string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res bool
	err := c.WS.SendRequestSync(ctx, "processResults", inventory, &res)
	if err != nil {
		return false, err
	}
	return res, nil
}
