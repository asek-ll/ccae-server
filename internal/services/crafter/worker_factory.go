package crafter

import (
	"sync"

	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type WorkerFactory struct {
	storage storage.ItemStore
	daos    *dao.DaoProvider

	workers map[string]*CraftWorker
	mu      sync.Mutex
}

func NewWorkerFactory(storage storage.ItemStore, daos *dao.DaoProvider) *WorkerFactory {
	return &WorkerFactory{
		storage: storage,
		daos:    daos,

		workers: make(map[string]*CraftWorker),
	}
}

func (f *WorkerFactory) NewWorker(id string, client *wsmethods.CrafterClient) *CraftWorker {
	worker := NewCraftWorker(id, client, f.storage, f.daos)

	f.mu.Lock()
	defer f.mu.Unlock()

	if previousWorker, e := f.workers[id]; e {
		previousWorker.Stop()
	}

	f.workers[id] = worker
	worker.Start()

	return worker
}

func (f *WorkerFactory) HandleClientConnected(client wsmethods.Client) {
	crafterClient, e := client.(*wsmethods.CrafterClient)
	if e {
		f.NewWorker(crafterClient.ID, crafterClient)
	}
}

func (f *WorkerFactory) HandleClientDisconnected(client wsmethods.Client) {
	crafterClient, e := client.(*wsmethods.CrafterClient)
	if e {
		f.workers[crafterClient.ID].Stop()
		f.mu.Lock()
		delete(f.workers, crafterClient.ID)
		f.mu.Unlock()
	}
}
