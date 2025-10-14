package ws

import (
	"encoding/json"
	"io"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Client struct {
	ID      uint
	conn    net.Conn
	io      sync.Mutex
	handler Handler
	ExtID   string
}

func (c *Client) Receive() error {
	msg, err := c.readMessage()
	if err != nil {
		c.conn.Close()
		return err
	}
	if msg == nil {
		// Handled some control message.
		return nil
	}

	return c.handler.HandleMessage(msg, c)
}

func (c *Client) readMessage() ([]byte, error) {
	c.io.Lock()
	defer c.io.Unlock()

	h, r, err := wsutil.NextReader(c.conn, ws.StateServerSide)
	if err != nil {
		return nil, err
	}
	if h.OpCode.IsControl() {
		return nil, wsutil.ControlFrameHandler(c.conn, ws.StateServerSide)(h, r)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) WriteJSON(x any) error {
	w := wsutil.NewWriter(c.conn, ws.StateServerSide, ws.OpText)
	encoder := json.NewEncoder(w)

	c.io.Lock()
	defer c.io.Unlock()

	if err := encoder.Encode(x); err != nil {
		return err
	}

	return w.Flush()
}

func (c *Client) WriteText(p []byte) error {
	w := wsutil.NewWriter(c.conn, ws.StateServerSide, ws.OpText)
	c.io.Lock()
	defer c.io.Unlock()

	_, err := w.Write(p)
	if err != nil {
		return err
	}

	return w.Flush()
}

func (c *Client) writeRaw(p []byte) error {
	c.io.Lock()
	defer c.io.Unlock()

	_, err := c.conn.Write(p)

	return err
}
