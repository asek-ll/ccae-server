package app

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/services/player"
	"github.com/asek-ll/aecc-server/internal/services/recipe"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/pkg/logger"
)

type App struct {
	Daos          *dao.DaoProvider
	Storage       *storage.Storage
	Planner       *crafter.Planner
	RecipeManager *recipe.RecipeManager
	PlayerManager *player.PlayerManager
	Logger        *logger.Logger
}
