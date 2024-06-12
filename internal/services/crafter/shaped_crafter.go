package crafter

import (
	"log"

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
	err := c.cleanUp()
	if err != nil {
		return err
	}
	for _, ing := range recipe.Ingredients {
		_, err := c.storage.ExportStack(ing.ItemUID, c.crafter.BufferName(), *ing.Slot, ing.Amount*repeats)
		if err != nil {
			c.cleanUp()
			return err
		}
	}

	success, err := c.crafter.Craft()
	if err != nil {
		return err
	}

	if success {
		_, err = c.crafter.DumpOut()
		log.Println("[WARN] CRAFT!!! SUCCESS!!")
	}

	return c.cleanUp()
}

func (c *ShapedCrafter) cleanUp() error {
	_, err := c.crafter.DumpOut()
	if err != nil {
		return err
	}
	return c.storage.ImportAll(c.crafter.BufferName())
}
