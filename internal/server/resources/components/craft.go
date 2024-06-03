package components

import "github.com/asek-ll/aecc-server/internal/dao"

type CraftItem struct {
	Craft  *dao.Craft
	Recipe *dao.Recipe
}
