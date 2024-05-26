package storage

import "github.com/asek-ll/aecc-server/internal/common"

type ItemRef struct {
	Name string `json:"name"`
	NBT  string `json:"nbt"`
}

type Stack struct {
	Name     string `json:"name"`
	NBT      string `json:"nbt"`
	Count    int    `json:"count"`
	MaxCount int    `json:"maxCount"`
}

func (s Stack) GetUID() string {
	var nbt *string
	if s.NBT != "" {
		nbt = &s.NBT
	}
	return common.MakeUid(s.Name, nbt)
}

type StackWithSlot struct {
	Item Stack `json:"item"`
	Slot int   `json:"slot"`
}

type Inventory struct {
	Name  string          `json:"name"`
	Items []StackWithSlot `json:"items"`
	Size  int             `json:"size"`
}

type SlotRef struct {
	InventoryName string
	Slot          int
}

type ExportParams struct {
	Item   ItemRef
	Target SlotRef
	Amount int
}
