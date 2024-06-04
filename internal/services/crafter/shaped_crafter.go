package crafter

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type ShapedCrafter struct {
	storage *wsmethods.StorageClient
	crafter *wsmethods.CrafterClient
}

func NewShapedCrafter(storage *wsmethods.StorageClient, crafter *wsmethods.CrafterClient) *ShapedCrafter {
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
		slot := wsmethods.SlotRef{
			InventoryName: c.crafter.BufferName(),
			Slot:          *ing.Slot,
		}
		err := c.storage.ExportStack([]wsmethods.ExportParams{
			{Item: wsmethods.ItemRefFromUid(ing.ItemUID), Target: slot, Amount: ing.Amount * repeats},
		})
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
