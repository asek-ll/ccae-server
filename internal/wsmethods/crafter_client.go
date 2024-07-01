package wsmethods

import (
	"context"
	"fmt"
	"time"
)

type RecipeDto struct {
	Type string `json:"type"`
}

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
		bufferName:    bufferName.(string),
	}, nil
}

func (c *CrafterClient) BufferName() string {
	return c.bufferName
}

func (c *CrafterClient) Craft(recipe RecipeDto) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res bool
	err := c.WS.SendRequestSync(ctx, "craft", recipe, &res)
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

func (c *CrafterClient) Restore(recipe RecipeDto) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res bool
	err := c.WS.SendRequestSync(ctx, "restore", recipe, &res)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (c *CrafterClient) GetSupportTypes() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var res []string
	err := c.WS.SendRequestSync(ctx, "getTypes", nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
