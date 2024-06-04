package wsmethods

import (
	"context"
	"fmt"
	"time"

	"github.com/asek-ll/aecc-server/internal/wsrpc"
)

type CrafterClient struct {
	GenericClient
	bufferName string
}

func NewCrafterClient(ID string, WS wsrpc.ClientWrapper, props map[string]any) (*CrafterClient, error) {
	bufferName, ok := props["buffer_name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid buffer_name: %v", props["buffer_name"])
	}
	return &CrafterClient{
		GenericClient: GenericClient{
			ID: ID,
			WS: WS,
		},
		bufferName: bufferName,
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
