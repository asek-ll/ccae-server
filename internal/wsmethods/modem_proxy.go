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

func (c *ModemAdapter) GetNamesRemote(ctx context.Context) ([]string, error) {
	return CallWithClientForType(c.clientsManager, func(client *ModemAdapter) ([]string, error) {
		return client.GetNamesRemote(ctx)
	})
}

func (c *ModemAdapter) GetMethodsRemote(ctx context.Context, name string) ([]string, error) {
	return CallWithClientForType(c.clientsManager, func(client *ModemAdapter) ([]string, error) {
		return client.GetMethodsRemote(ctx, name)
	})
}
