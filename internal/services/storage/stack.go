package storage

type Inventory struct {
	ID string
}

type InventorySlot struct {
	InvetoryId string
	SlotNumber int
}

type ItemStack struct {
	InventorySlot
	Name     string
	NBT      string
	Count    int
	MaxCount int
}

type StorageItem struct {
	Name     string
	NBT      string
	Count    int
	MaxCount int
	Stacks   []ItemStack
}
