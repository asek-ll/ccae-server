package ws

type Handler interface {
	HandleMessage(content []byte, client *Client) error
	HandleDisconnect(client *Client)
}

type delegateHandler struct {
	delegate Handler
}

func (d *delegateHandler) HandleMessage(content []byte, client *Client) error {
	if d.delegate != nil {
		return d.delegate.HandleMessage(content, client)
	}
	return nil
}

func (d *delegateHandler) HandleDisconnect(client *Client) {
	if d.delegate != nil {
		d.delegate.HandleDisconnect(client)
	}
}
