package crafter

import "github.com/asek-ll/aecc-server/internal/dao"

type Step struct {
	Recipe  *dao.Recipe
	Repeats int
}

type Related struct {
	UID           string
	Produced      int
	Consumed      int
	StorageAmount int
}

type Plan struct {
	Items   []string
	Steps   []Step
	Related []*Related
	Goals   []dao.RecipeItem
}
