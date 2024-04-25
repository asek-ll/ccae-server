package cmd

import (
	"net/http"
	"time"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/server"
	"github.com/asek-ll/aecc-server/internal/ws"
	"github.com/asek-ll/aecc-server/internal/wshandler"
	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = ServerCommand{}

type ServerCommand struct {
}

func (s ServerCommand) Execute(args []string) error {

	app, err := app.CreateApp()
	if err != nil {
		return err
	}

	mux, err := server.CreateMux(app)
	if err != nil {
		return err
	}

	wshandler := &wshandler.Handler{}
	server := ws.NewServer(":12526", 128, 1, time.Millisecond*1000, wshandler)

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
		err := server.Start()
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
