package crafter

import "github.com/asek-ll/aecc-server/internal/dao"

type Step struct {
	Recipe  *dao.Recipe
	Repeats int
}

type Related struct {
	Produced      int
	Consumed      int
	StorageAmount int
}

type Plan struct {
	Items   []string
	Steps   []Step
	Goals   []Stack
	Related map[string]*Related
}
