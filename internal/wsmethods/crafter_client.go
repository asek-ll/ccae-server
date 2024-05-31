package wsmethods

import "time"

func RepeatUntil(period time.Duration, proc func(), end chan struct{}) {
	for {
		select {
		case <-time.After(period):
			proc()
		case <-end:
			return
		}
	}
}

type CrafterClient struct {
	clientId int
	end      chan struct{}
}

func (c *CrafterClient) process() {
	//select craft
}

func (c *CrafterClient) Start() {
	go RepeatUntil(10*time.Second, func() {
		c.process()
	}, c.end)
}

func (c *CrafterClient) Stop() {
	c.end <- struct{}{}
}
