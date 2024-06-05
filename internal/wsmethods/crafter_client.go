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
	bufferName, ok := base.Props["buffer_name"].(string)
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

func (c *CrafterClient) Craft() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return c.WS.SendRequestSync(ctx, "craft", nil, nil)
}

func (c *CrafterClient) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return c.WS.SendRequestSync(ctx, "cleanup", nil, nil)
}
