package wsmethods

import "context"

type ModemAdapter struct {
	clientsManager *ClientsManager
}

func NewModemAdapter(clientsManager *ClientsManager) *ModemAdapter {
	return &ModemAdapter{
		clientsManager: clientsManager,
	}
}

func (c *ModemAdapter) GetNamesRemote() ([]string, error) {
	return CallWithClientForType(c.clientsManager, func(client *ModemClient) ([]string, error) {
		return client.GetNamesRemote()
	})
}

func (c *ModemAdapter) GetMethodsRemote(ctx context.Context, name string) ([]string, error) {
	return CallWithClientForType(c.clientsManager, func(client *ModemClient) ([]string, error) {
		return client.GetMethodsRemote(ctx, name)
	})
}
