package crafter

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type ShapedCrafter struct {
	storage storage.ItemStore
	crafter *wsmethods.CrafterClient
}

func NewShapedCrafter(storage storage.ItemStore, crafter *wsmethods.CrafterClient) *ShapedCrafter {
	return &ShapedCrafter{
		storage: storage,
		crafter: crafter,
	}
}

func (c *ShapedCrafter) Craft(recipe *dao.Recipe, repeats int) error {
	err := c.crafter.Cleanup()
	if err != nil {
		return err
	}
	for _, ing := range recipe.Ingredients {
		_, err := c.storage.ExportStack(ing.ItemUID, c.crafter.BufferName(), *ing.Slot, ing.Amount*repeats)
		if err != nil {
			c.crafter.Cleanup()
			return err
		}
	}

	err = c.crafter.Craft()
	if err != nil {
		return err
	}

	return c.crafter.Cleanup()
}
