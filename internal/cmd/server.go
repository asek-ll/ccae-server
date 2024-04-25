package cmd

import (
	"net/http"

	"github.com/asek-ll/aecc-server/internal/app"
	"github.com/asek-ll/aecc-server/internal/server"
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

	errors := make(chan error)

	go func() {
		err := http.ListenAndServe(":3001", mux)
		if err != nil {
			errors <- err
		}
	}()

	return <-errors
}
