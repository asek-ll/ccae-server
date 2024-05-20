package app

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/services/storage"
)

type App struct {
	Daos    *dao.DaoProvider
	Storage *storage.Storage
	Planner *crafter.Planner
}
