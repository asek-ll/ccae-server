package storage

type Storage interface {
	GetInventories() []Inventory
	GetItems() []StorageItem
	Transfer(from InventorySlot, to InventorySlot, amount int)
}

func Refresh() {

}

func Transfer() {

}
