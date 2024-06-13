package wsmethods

type ClientListener interface {
	HandleClientConnected(client Client)
	HandleClientDisconnected(client Client)
}

type DumpCycleListener struct {
}

func (l DumpCycleListener) HandleClientConnected(client Client) {
}

func (l DumpCycleListener) HandleClientDisconnected(client Client) {
}
