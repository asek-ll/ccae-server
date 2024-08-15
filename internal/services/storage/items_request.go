package storage

type ExportRequest struct {
	RequestItems  []ExportRequestItems
	RequestFluids []ExportRequestFluids
}
type ExportRequestItems struct {
	TargetStorage string
	Uid           string
	ToSlot        int
	Amount        int
}

type ExportRequestFluids struct {
	TargetTankName string
	Uid            string
	Amount         int
}

type ExportTransactionData struct {
	ItemStacks  []ExportTransactionStorageSlot
	FluidStacks []ExportTransactionTank
}

type ExportTransactionStorageSlot struct {
	StorageName   string
	Slot          int
	TargetStorage string
	Uid           string
	ToSlot        int
	Amount        int
}

type ExportTransactionTank struct {
	TankName       string
	TargetTankName string
	Uid            string
	Amount         int
}
