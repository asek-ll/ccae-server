package cmd

import (
	"fmt"
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
	"github.com/asek-ll/aecc-server/pkg/logger"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = ServerCommand{}

type ServerCommand struct {
}

func (s ServerCommand) Execute(args []string) error {

	logfmt := logger.Format(
		fmt.Sprintf("%s [{{.Level}}] {{.Message}}",
			color.GreenString(`{{.Time.Format "2006-01-02T15:04:05"}}`),
		),
	)
	l := logger.New(logfmt, logger.LevelFormat(func(l logger.Level) string {
		if l.Value >= logger.ERROR.Value {
			return color.RedString(l.Padded())
		}
		if l.Value >= logger.WARN.Value {
			return color.HiRedString(l.Padded())
		}
		if l.Value >= logger.INFO.Value {
			return l.Padded()
		}
		return color.GreenString(l.Padded())
	}), logger.WithLevel(logger.TRACE),
	)

	logger.SetupStd(l)

	daos, err := dao.NewDaoProvider()
	if err != nil {
		return err
	}

	wsServer := ws.NewServer(":12526", 128, 1, time.Millisecond*1000)
	rpcServer := wsrpc.NewServer(wsServer)

	storageService := storage.NewStorage(rpcServer, daos)
	playerManager := player.NewPlayerManager(rpcServer, daos)
	plannerService := crafter.NewPlanner(daos, storageService)
	recipeManager := recipe.NewRecipeManager(daos)

	app := &app.App{
		Daos:          daos,
		Storage:       storageService,
		Planner:       plannerService,
		RecipeManager: recipeManager,
		PlayerManager: playerManager,
		Logger:        log.Default(),
	}

	mux, err := server.CreateMux(app)
	if err != nil {
		return err
	}

	wsmethods.SetupMethods(rpcServer, app)

	l.Logf(logger.TRACE, "Ready")
	l.Logf(logger.DEBUG, "Steady")
	l.Logf(logger.INFO, "Go")
	l.Logf(logger.WARN, "!!!")
	l.Logf(logger.ERROR, "Fail")

	errors := make(chan error)
	done := make(chan struct{})

	go func() {
		l.Logf(logger.INFO, "Start http server")
		err := http.ListenAndServe(":3001", mux)
		if err != nil {
			errors <- err
		} else {
			done <- struct{}{}
		}
	}()

	go func() {
		l.Logf(logger.INFO, "Start websocket server")
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
