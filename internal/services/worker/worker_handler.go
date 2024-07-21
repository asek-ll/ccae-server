package worker

import (
	"log"
	"time"
)

type WorkerHandler struct {
	key       string
	runner    func() error
	isStarted bool
	done      chan bool
	ping      chan bool
}

func (wh *WorkerHandler) IsStarted() bool {
	return wh.isStarted
}
func (c *WorkerHandler) Stop() {
	if !c.isStarted {
		return
	}
	log.Printf("[INFO] Stop worker for '%s'", c.key)
	c.isStarted = false
	c.done <- true
}

func (c *WorkerHandler) Start() error {
	log.Printf("[INFO] Start worker for '%s'", c.key)
	c.isStarted = true
	go c.work()
	return nil
}

func (c *WorkerHandler) Ping() {
	log.Printf("[INFO] Ping to %s", c.key)
	select {
	case c.ping <- true:
		return
	default:
		return
	}
}

func (wh *WorkerHandler) work() {
	for {
		if !wh.isStarted {
			<-wh.done
			return
		}
		err := wh.runner()
		if err != nil {
			log.Printf("[ERROR] Can't process worker '%s', error: %v", wh.key, err)
		}
		select {
		case <-wh.done:
			return
		case <-time.After(time.Second * 30):
			continue
		case <-wh.ping:
			continue
		}
	}
}
