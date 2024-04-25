package ws

type Handler interface {
	HandleMessage(content []byte, clientId uint) error
}
