package crafter

type Recipe struct {
	Type        string
	Input       map[int]Stack
	Output      []Stack
	MaxBachSize int
}
