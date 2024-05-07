package crafter

import "github.com/asek-ll/aecc-server/internal/dao"

type Stack struct {
	ItemId dao.ItemId
	Count  int
}
