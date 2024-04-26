package ws

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/asek-ll/aecc-server/pkg/gopool"
	"github.com/gobwas/ws"
	"github.com/mailru/easygo/netpoll"
)

type Server struct {
	addr string
	pool *gopool.Pool

	clients    map[uint]*Client
	clientsMu  sync.RWMutex
	clientsSec uint

	exit      chan struct{}
	ioTimeout time.Duration

	handler *delegateHandler
}

func NewServer(addr string, workers int, queue int, ioTimeout time.Duration) *Server {
	pool := gopool.NewPool(workers, queue, 1)
	handler := delegateHandler{}
	exit := make(chan struct{})

	return &Server{
		addr:      addr,
		pool:      pool,
		clients:   make(map[uint]*Client),
		exit:      exit,
		ioTimeout: ioTimeout,
		handler:   &handler,
	}
}

func (s *Server) SetHandler(handler Handler) {
	s.handler.delegate = handler
}

func (s *Server) Start() error {
	poller, err := netpoll.New(nil)
	if err != nil {
		return err
	}
	handle := func(conn net.Conn) {
		safeConn := NewDeadliner(conn, s.ioTimeout)

		// Zero-copy upgrade to WebSocket connection.
		hs, err := ws.Upgrade(safeConn)
		if err != nil {
			log.Printf("%s: upgrade error: %v", nameConn(conn), err)
			conn.Close()
			return
		}

		log.Printf("%s: established websocket connection: %+v", nameConn(conn), hs)

		client := s.register(safeConn)

		desc := netpoll.Must(netpoll.HandleReadOnce(conn))

		poller.Start(desc, func(ev netpoll.Event) {
			if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
				poller.Stop(desc)
				s.remove(client)
				return
			}
			s.pool.Schedule(func() {
				if err := client.Receive(); err != nil {
					// When receive failed, we can only disconnect broken
					// connection and stop to receive events about it.
					poller.Stop(desc)
					s.remove(client)
				} else {
					poller.Resume(desc)
				}
			})
		})
	}

	// Create incoming connections listener.
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Printf("websocket is listening on %s", ln.Addr().String())

	// Create netpoll descriptor for the listener.
	// We use OneShot here to manually resume events stream when we want to.
	acceptDesc := netpoll.Must(netpoll.HandleListener(
		ln, netpoll.EventRead|netpoll.EventOneShot,
	))

	// accept is a channel to signal about next incoming connection Accept()
	// results.
	accept := make(chan error, 1)

	// Subscribe to events about listener.
	poller.Start(acceptDesc, func(e netpoll.Event) {
		// We do not want to accept incoming connection when goroutine pool is
		// busy. So if there are no free goroutines during 1ms we want to
		// cooldown the server and do not receive connection for some short
		// time.
		err := s.pool.ScheduleTimeout(time.Millisecond, func() {
			conn, err := ln.Accept()
			if err != nil {
				accept <- err
				return
			}

			accept <- nil
			handle(conn)
		})
		if err == nil {
			err = <-accept
		}
		if err != nil {
			if err != gopool.ErrScheduleTimeout {
				goto cooldown
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				goto cooldown
			}

			log.Fatalf("accept error: %v", err)

		cooldown:
			delay := 5 * time.Millisecond
			log.Printf("accept error: %v; retrying in %s", err, delay)
			time.Sleep(delay)
		}

		poller.Resume(acceptDesc)
	})

	<-s.exit

	return nil
}

func (c *Server) register(conn net.Conn) *Client {
	client := &Client{
		conn:    conn,
		handler: c.handler,
	}
	c.clientsMu.Lock()
	{
		client.id = c.clientsSec
		c.clients[client.id] = client
		c.clientsSec++
	}
	c.clientsMu.Unlock()

	return client
}

func (c *Server) remove(client *Client) {
	c.clientsMu.Lock()
	if _, e := c.clients[client.id]; e {
		delete(c.clients, client.id)
	}
	c.clientsMu.Unlock()

	c.handler.HandleDisconnect(client)
}

func (c *Server) GetClient(id uint) (*Client, bool) {
	c.clientsMu.RLock()
	defer c.clientsMu.RUnlock()
	client, e := c.clients[id]
	return client, e
}

func nameConn(conn net.Conn) string {
	return conn.LocalAddr().String() + " > " + conn.RemoteAddr().String()
}
