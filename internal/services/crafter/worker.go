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
	storage  *storage.Storage
	daos     *dao.DaoProvider

	stopped bool
	done    chan bool

	ping chan bool
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
	select {
	case c.ping <- true:
		return
	default:
		return
	}
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
		} else {
			// err = c.processResults()
			// if err != nil {
			// log.Printf("[ERROR] Can't process results for worker '%s', error: %v", c.workerId, err)
			// } else
			if done {
				continue
			}
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

// func (c *CraftWorker) processResults() error {
// 	if c.resultProcessedTime.Add(time.Second * 30).After(time.Now()) {
// 		return nil
// 	}
// 	input, err := c.storage.GetInput()
// 	if err != nil {
// 		return err
// 	}
// 	c.resultProcessedTime = time.Now()

// 	_, err = c.client.ProcessResults(input)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (c *CraftWorker) process() (bool, error) {
	current, err := c.daos.Crafts.FindCurrent(c.workerId)
	if err != nil {
		return false, err
	}
	if current != nil {
		recipe, err := c.daos.Recipes.GetRecipeById(current.RecipeID)
		if err != nil {
			return false, err
		}

		done, err := c.client.Restore(wsmethods.RecipeDto{Type: recipe.Type})
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
	repeats := min(next.Repeats, maxRepeats)

	err = c.trasferItems(recipe, repeats)
	if err != nil {
		return false, err
	}

	err = c.daos.Crafts.CommitCraft(next, recipe, repeats)
	if err != nil {
		return false, err
	}
	done, err := c.client.Craft(wsmethods.RecipeDto{Type: recipe.Type})
	if done {
		err = c.daos.Crafts.CompleteCraft(next)
		if err != nil {
			return false, err
		}
	} else {
		err = c.daos.Crafts.SuspendCraft(next, recipe)
		if err != nil {
			return false, err
		}
	}
	return true, nil

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
