package worker

import "sync"

type WorkerHandlerManager struct {
	workerHanlders map[string]*WorkerHandler

	mu sync.RWMutex
}

func NewWorkerHandlerManager() *WorkerHandlerManager {
	return &WorkerHandlerManager{
		workerHanlders: make(map[string]*WorkerHandler),
	}
}

func (w *WorkerHandlerManager) Ping(key string) {
	w.mu.RLock()
	if h, e := w.workerHanlders[key]; e {
		h.Ping()
	}
	w.mu.RUnlock()
}

func (w *WorkerHandlerManager) Add(key string, runner func() error) (*WorkerHandler, error) {
	w.mu.Lock()
	if h, e := w.workerHanlders[key]; e {
		h.Stop()
	}
	worker := &WorkerHandler{
		key:       key,
		runner:    runner,
		isStarted: false,
		done:      make(chan bool),
		ping:      make(chan bool),
	}
	w.workerHanlders[key] = worker
	w.mu.Unlock()
	return worker, nil
}

func (w *WorkerHandlerManager) Remove(key string) {
	w.mu.Lock()
	if h, e := w.workerHanlders[key]; e {
		h.Stop()
	}
	delete(w.workerHanlders, key)
	w.mu.Unlock()
}
