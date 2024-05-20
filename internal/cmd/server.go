package cmd

import (
	"net/http"
	"time"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/dao"
	"github.com/asek-ll/aecc-server/internal/server"
	"github.com/asek-ll/aecc-server/internal/services/crafter"
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

	daos, err := dao.NewDaoProvider()
	if err != nil {
		return err
	}

	wsServer := ws.NewServer(":12526", 128, 1, time.Millisecond*1000)
	rpcServer := wsrpc.NewServer(wsServer)

	storageService := storage.NewStorage(rpcServer, daos)
	plannerService := crafter.NewPlanner(daos, storageService)

	app := &app.App{
		Daos:    daos,
		Storage: storageService,
		Planner: plannerService,
	}

	mux, err := server.CreateMux(app)
	if err != nil {
		return err
	}

	wsmethods.SetupMethods(rpcServer, app)

	errors := make(chan error)
	done := make(chan struct{})

	go func() {
		err := http.ListenAndServe(":3001", mux)
		if err != nil {
			errors <- err
		} else {
			done <- struct{}{}
		}
	}()

	go func() {
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
