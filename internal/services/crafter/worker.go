package crafter

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type CraftWorker struct {
	workerId string
	client   *wsmethods.CrafterClient
	storage  *storage.Storage
	daos     *dao.DaoProvider

	done chan bool
	ping chan bool

	lastTypes []string
	stopped   bool
}

func NewCraftWorker(
	workerId string,
	client *wsmethods.CrafterClient,
	storage *storage.Storage,
	daos *dao.DaoProvider,
) *CraftWorker {
	return &CraftWorker{
		workerId: workerId,
		storage:  storage,
		daos:     daos,
		client:   client,

		done: make(chan bool),
		ping: make(chan bool, 1),
	}
}

func (c *CraftWorker) Stop() {
	if c.stopped {
		return
	}
	log.Printf("[INFO] Stop worker for '%s'", c.workerId)
	c.stopped = true
	c.done <- true
}

func (c *CraftWorker) Start() error {
	log.Printf("[INFO] Start worker for '%s'", c.workerId)
	go c.cycle()
	return nil
}

func (c *CraftWorker) Ping() {
	log.Printf("[INFO] Ping to %s", c.workerId)
	select {
	case c.ping <- true:
		return
	default:
		return
	}
}

func (c *CraftWorker) GetLastTypes() []string {
	return c.lastTypes
}

func (c *CraftWorker) cycle() {
	for {
		if c.stopped {
			<-c.done
			return
		}
		err := c.process()
		if err != nil {
			log.Printf("[ERROR] Can't process craft for worker '%s', error: %v", c.workerId, err)
		}
		select {
		case <-c.done:
			return
		case <-time.After(time.Second * 30):
			continue
		case <-c.ping:
			continue
		}
	}
}

func (c *CraftWorker) Restore(craft *dao.Craft) (bool, error) {
	if craft.Status == dao.PENDING_CRAFT_STATUS {
		return c.Craft(craft)
	}

	recipe, err := c.daos.Recipes.GetRecipeById(craft.RecipeID)
	if err != nil {
		return false, err
	}

	done, err := c.client.Restore(wsmethods.RecipeDto{Type: recipe.Type})
	if err != nil {
		return false, err
	}
	if done {
		err = c.daos.Crafts.CompleteCraft(craft)
		if err != nil {
			return false, err
		}
	}
	return done, nil
}

func (c *CraftWorker) Craft(craft *dao.Craft) (bool, error) {
	cleaned, err := c.client.DumpOut()
	if err != nil {
		return false, err
	}
	if !cleaned {
		return false, errors.New("failed to dump out")
	}

	err = c.cleanupBuffer()
	if err != nil {
		return false, err
	}

	recipe, err := c.daos.Recipes.GetRecipeById(craft.RecipeID)
	if err != nil {
		return false, err
	}

	maxIngs := 1
	for _, r := range recipe.Results {
		maxIngs = max(maxIngs, r.Amount)
	}
	for _, r := range recipe.Ingredients {
		maxIngs = max(maxIngs, r.Amount)
	}

	maxRepeats := 64 / maxIngs
	if maxRepeats == 0 {
		return false, errors.New("Illegal max repeats for recipe")
	}

	if recipe.MaxRepeats != nil {
		maxRepeats = *recipe.MaxRepeats
	}
	repeats := min(craft.Repeats, maxRepeats)

	err = c.trasferItems(recipe, repeats)
	if err != nil {
		return false, err
	}

	err = c.daos.Crafts.CommitCraft(craft, recipe, repeats)
	if err != nil {
		return false, err
	}
	done, err := c.client.Craft(wsmethods.RecipeDto{Type: recipe.Type})
	if done {
		err = c.daos.Crafts.CompleteCraft(craft)
		if err != nil {
			return false, err
		}
	}
	return done, nil
}

func (c *CraftWorker) process() error {
	current, err := c.daos.Crafts.FindCurrent(c.workerId)
	if err != nil {
		return fmt.Errorf("Can't find current craft: %w", err)
	}
	if current != nil {
		done, err := c.Restore(current)
		if err != nil {
			return fmt.Errorf("Can't restore current craft: %w", err)
		}
		if !done {
			return nil
		}
	}

	types, err := c.client.GetSupportTypes()
	if err != nil {
		return err
	}
	c.lastTypes = types

	if len(types) == 0 {
		return nil
	}

	active := true
	for active {
		active = false

		nexts, err := c.daos.Crafts.FindNextByTypes(types, c.workerId)
		if err != nil {
			return err
		}

		log.Println(types, nexts)

		for _, craft := range nexts {
			assigned, err := c.daos.Crafts.AssignCraftToWorker(craft, c.workerId)
			if err != nil {
				return err
			}
			if assigned {
				done, err := c.Craft(craft)
				if err != nil {
					return err
				}
				if !done {
					return nil
				}
				active = true
			}
		}
	}

	return nil

}
func (c *CraftWorker) trasferItems(recipe *dao.Recipe, repeats int) error {
	log.Printf("[INFO] Transfer items for '%s' and recipe: %v", c.workerId, recipe)
	for i, ing := range recipe.Ingredients {
		slot := i + 1
		if ing.Slot != nil {
			slot = *ing.Slot
		}
		_, err := c.storage.ExportStack(ing.ItemUID, c.client.BufferName(), slot, ing.Amount*repeats)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CraftWorker) cleanupBuffer() error {
	return c.storage.ImportAll(c.client.BufferName())
}
