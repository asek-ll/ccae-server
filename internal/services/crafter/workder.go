package crafter

import (
	"github.com/asek-ll/aecc-server/internal/dao"
)

type Worker interface {
	Craft(recipe *dao.Recipe, repeats int) error
}
