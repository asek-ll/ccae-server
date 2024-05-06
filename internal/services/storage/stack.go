package storage

type StackId struct {
	Name string `json:"name"`
	NBT  string `json:"nbt"`
}

type Stack struct {
	Name     string `json:"name"`
	NBT      string `json:"nbt"`
	Count    int    `json:"count"`
	MaxCount int    `json:"maxCount"`
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
