package cmd

import (
	"log"
	"net/http"
	"time"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/config"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/server"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/services/modem"
	"github.com/asek-ll/aecc-server/internal/services/player"
	"github.com/asek-ll/aecc-server/internal/services/recipe"
	"github.com/asek-ll/aecc-server/internal/services/storage"
	"github.com/asek-ll/aecc-server/internal/services/worker"
	"github.com/asek-ll/aecc-server/internal/ws"
	"github.com/asek-ll/aecc-server/internal/wsmethods"
	"github.com/asek-ll/aecc-server/internal/wsrpc"
	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = ServerCommand{}

type ServerCommand struct {
}

func (s ServerCommand) Execute(args []string) error {

	l := setupLogger()

	configLoader, err := config.NewConfigLoader()
	if err != nil {
		return err
	}

	daos, err := dao.NewDaoProvider()
	if err != nil {
		return err
	}

	wsServer := ws.NewServer(":12526", 128, 1, time.Millisecond*1000)
	rpcServer := wsrpc.NewServer(wsServer)

	clientsManager := wsmethods.NewClientsManager(rpcServer, daos.Clients)
	storageAdapter := wsmethods.NewStorageAdapter(clientsManager)

	storageService := storage.NewStorage(daos, storageAdapter)
	playerManager := player.NewPlayerManager(daos, clientsManager, storageService)
	plannerService := crafter.NewPlanner(daos, storageService)
	recipeManager := recipe.NewRecipeManager(daos)
	workerFactory := crafter.NewWorkerFactory(storageService, daos)
	crafterService := crafter.NewCrafter(daos, plannerService, workerFactory, storageService)
	modemManager := modem.NewModemManager(clientsManager)
	transferTransationManager := storage.NewTransferTransactionManager(configLoader, daos.StoredTX, storageAdapter, storageService)

	stateUpdater := crafter.NewStateUpdater(storageService, daos, crafterService)
	stateUpdater.Start()

	exporterWorker := worker.NewExporterWorker(*storageService)
	importerWorker := worker.NewImporterWorker(*storageService)
	processingCrafterWorker := worker.NewProcessingCrafterWorker(daos, storageService, storageAdapter, transferTransationManager)
	workerManager := worker.NewWorkerManager(configLoader, daos, exporterWorker, importerWorker, processingCrafterWorker)

	clientsManager.SetClientListener(workerFactory)

	app := &app.App{
		Daos:                       daos,
		Storage:                    storageService,
		Planner:                    plannerService,
		Crafter:                    crafterService,
		RecipeManager:              recipeManager,
		PlayerManager:              playerManager,
		Logger:                     log.Default(),
		ClientsManager:             clientsManager,
		WorkerFactory:              workerFactory,
		WorkerManager:              workerManager,
		ModemManager:               modemManager,
		StorageAdapter:             storageAdapter,
		TransferTransactionManager: transferTransationManager,
		ConfigLoader:               configLoader,
	}

	mux, err := server.CreateMux(app)
	if err != nil {
		return err
	}

	errors := make(chan error)
	done := make(chan struct{})

	go func() {
		l.Println("INFO Start http server")
		err := http.ListenAndServe(":3001", mux)
		if err != nil {
			errors <- err
		} else {
			done <- struct{}{}
		}
	}()

	go func() {
		l.Println("INFO Start websocket server")
		err := wsServer.Start()
		if err != nil {
			errors <- err
		} else {
			done <- struct{}{}
		}
	}()

	select {
	case errors <- err:
		return err
	case <-done:
		return nil
	}
}
