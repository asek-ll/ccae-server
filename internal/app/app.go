package app

import (
	"log"

	"github.com/asek-ll/aecc-server/internal/config"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/services/clientscripts"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/services/item"
	"github.com/asek-ll/aecc-server/internal/services/modem"
	"github.com/asek-ll/aecc-server/internal/services/player"
	"github.com/asek-ll/aecc-server/internal/services/recipe"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/services/worker"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
)

type App struct {
	Daos                       *dao.DaoProvider
	Storage                    *storage.Storage
	Planner                    *crafter.Planner
	Crafter                    *crafter.Crafter
	RecipeManager              *recipe.RecipeManager
	PlayerManager              *player.PlayerManager
	Logger                     *log.Logger
	ClientsManager             *wsmethods.ClientsManager
	WorkerFactory              *crafter.WorkerFactory
	WorkerManager              *worker.WorkerManager
	ModemManager               *modem.ModemManager
	StorageAdapter             *wsmethods.StorageAdapter
	TransferTransactionManager *storage.TransferTransactionManager
	ConfigLoader               *config.ConfigLoader
	ItemManager                *item.ItemManager
	ScriptsManager             *clientscripts.ScriptsManager
}
