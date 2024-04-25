package ws

import (
	"net"
	"sync"

	"github.com/asek-ll/aecc-server/pkg/gopool"
)

type Clients struct {
	cs      map[uint]*Client
	mu      sync.RWMutex
	pool    *gopool.Pool
	seq     uint
	handler Handler
}

func NewClients(pool *gopool.Pool, handler Handler) *Clients {
	return &Clients{
		pool:    pool,
		cs:      make(map[uint]*Client),
		handler: handler,
	}
}

func (c *Clients) Register(conn net.Conn) *Client {
	client := &Client{
		conn:    conn,
		handler: c.handler,
	}
	c.mu.Lock()
	{
		client.id = c.seq
		c.cs[client.id] = client
		c.seq++
	}
	c.mu.Unlock()

	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	names := []string{}
	// 	client.CallSync("getFileNames", map[string]interface{}{
	// 		"server": "home",
	// 	}, &names)

	// 	fmt.Println(names)
	// }()

	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	client.writeNotice(
	// 		"getFileNames",
	// 		map[string]interface{}{
	// 			"server": "home",
	// 		},
	// 	)

	// 	time.Sleep(2 * time.Second)
	// 	client.writeNotice(
	// 		"getFileNames",
	// 		map[string]interface{}{
	// 			"server": "home",
	// 		},
	// 	)

	// }()

	return client
}

func (c *Clients) Remove(client *Client) {
	c.mu.Lock()
	if _, e := c.cs[client.id]; e {
		delete(c.cs, client.id)
	}
	c.mu.Unlock()
}
