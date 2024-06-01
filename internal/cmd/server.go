package cmd

import (
	"log"
	"net/http"
	"time"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/server"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
	"github.com/asek-ll/aecc-server/internal/services/player"
	"github.com/asek-ll/aecc-server/internal/services/recipe"
	"github.com/asek-ll/aecc-server/internal/services/storage"
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

	daos, err := dao.NewDaoProvider()
	if err != nil {
		return err
	}

	wsServer := ws.NewServer(":12526", 128, 1, time.Millisecond*1000)
	rpcServer := wsrpc.NewServer(wsServer)

	storageService := storage.NewStorage(rpcServer, daos)
	playerManager := player.NewPlayerManager(rpcServer, daos)
	plannerService := crafter.NewPlanner(daos, storageService)
	crafterService := crafter.NewCrafter(daos, plannerService)
	recipeManager := recipe.NewRecipeManager(daos)

	app := &app.App{
		Daos:          daos,
		Storage:       storageService,
		Planner:       plannerService,
		Crafter:       crafterService,
		RecipeManager: recipeManager,
		PlayerManager: playerManager,
		Logger:        log.Default(),
	}

	mux, err := server.CreateMux(app)
	if err != nil {
		return err
	}

	wsmethods.SetupMethods(rpcServer, app)

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
