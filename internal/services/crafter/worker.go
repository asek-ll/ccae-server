package crafter

import (
	"errors"
	"log"
	"time"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type CraftWorker struct {
	workerId string
	client   *wsmethods.CrafterClient
	storage  storage.ItemStore
	daos     *dao.DaoProvider

	stopped bool
	done    chan bool
}

func NewCraftWorker(
	workerId string,
	client *wsmethods.CrafterClient,
	storage storage.ItemStore,
	daos *dao.DaoProvider,
) *CraftWorker {
	return &CraftWorker{
		workerId: workerId,
		storage:  storage,
		daos:     daos,
		client:   client,

		done: make(chan bool),
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

func (c *CraftWorker) cycle() {
	for {
		if c.stopped {
			<-c.done
			return
		}
		done, err := c.process()
		if err != nil {
			log.Printf("[ERROR] Can't process craft for worker '%s', error: %v", c.workerId, err)
		} else if done {
			continue
		}
		select {
		case <-c.done:
			return
		case <-time.After(time.Second * 30):
			continue
		}
	}
}

func (c *CraftWorker) process() (bool, error) {
	current, err := c.daos.Crafts.FindCurrent(c.workerId)
	if err != nil {
		return false, err
	}
	if current != nil {
		done, err := c.client.Restore()
		if err != nil {
			return false, err
		}
		if done {
			err = c.daos.Crafts.CompleteCraft(current)
			if err != nil {
				return false, err
			}
		}
		return done, nil
	}

	next, err := c.daos.Crafts.FindNext(c.workerId)
	if err != nil {
		return false, err
	}
	if next == nil {
		return false, nil
	}

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

	recipe, err := c.daos.Recipes.GetRecipeById(next.RecipeID)
	if err != nil {
		return false, err
	}

	err = c.trasferItems(next, recipe)
	if err != nil {
		return false, err
	}

	err = c.daos.Crafts.CommitCraft(next, recipe)
	if err != nil {
		return false, err
	}
	done, err := c.client.Craft()
	if done {
		err = c.daos.Crafts.CompleteCraft(current)
		if err != nil {
			return false, err
		}
	}
	return done, nil

}
func (c *CraftWorker) trasferItems(craft *dao.Craft, recipe *dao.Recipe) error {
	for _, ing := range recipe.Ingredients {
		_, err := c.storage.ExportStack(ing.ItemUID, c.client.BufferName(), *ing.Slot, ing.Amount*craft.Repeats)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CraftWorker) cleanupBuffer() error {
	return c.storage.ImportAll(c.client.BufferName())
}

// func (c *CraftWorker) dumpOut() error {
// 	return nil
// }

// func (c *CraftWorker) restore() (bool, error) {
// 	return false, nil
// }

// func (c *CraftWorker) craft() (bool, error) {
// 	return false, nil
// }

// func (c *CraftWorker) bufferName() string {
// 	return ""
// }
