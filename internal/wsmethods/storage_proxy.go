package wsmethods

type StorageAdapter struct {
	clientsManager *ClientsManager
}

func NewStorageAdapter(clientsManager *ClientsManager) *StorageAdapter {
	return &StorageAdapter{
		clientsManager: clientsManager,
	}
}

func (s *StorageAdapter) GetProps() (map[string]string, error) {
	return CallWithClientForType(s.clientsManager, func(client *StorageClient) (map[string]string, error) {
		return client.Props, nil
	})
}

func (s *StorageAdapter) MoveStack(fromInventory string, fromSlot int, toInventory string, toSlot int, amount int) (int, error) {
	return CallWithClientForType(s.clientsManager, func(client *StorageClient) (int, error) {
		return client.MoveStack(fromInventory, fromSlot, toInventory, toSlot, amount)
	})
}

func (s *StorageAdapter) GetStackDetail(slotRef SlotRef) (*StackDetail, error) {
	return CallWithClientForType(s.clientsManager, func(client *StorageClient) (*StackDetail, error) {
		return client.GetStackDetail(slotRef)
	})
}

func (s *StorageAdapter) GetItems(prefixes []string) ([]Inventory, error) {
	return CallWithClientForType(s.clientsManager, func(client *StorageClient) ([]Inventory, error) {
		return client.GetItems(prefixes)
	})
}

func (s *StorageAdapter) ListItems(inventoryName string) ([]StackWithSlot, error) {
	return CallWithClientForType(s.clientsManager, func(client *StorageClient) ([]StackWithSlot, error) {
		return client.ListItems(inventoryName)
	})
}
