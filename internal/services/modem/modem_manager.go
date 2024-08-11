package modem

import "github.com/asek-ll/aecc-server/internal/wsmethods"

type ModemManager struct {
	modemAdapter *wsmethods.ModemAdapter
}

func NewModemManager(clientsManager *wsmethods.ClientsManager) *ModemManager {
	adapter := wsmethods.NewModemAdapter(clientsManager)
	return &ModemManager{
		modemAdapter: adapter,
	}
}

func (m *ModemManager) GetPeripherals() ([]string, error) {
	return m.modemAdapter.GetNamesRemote()
}
