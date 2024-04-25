package cmd

import (
	"net/http"

	"github.com/asek-ll/aecc-server/internal/server"
	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = ServerCommand{}

type ServerCommand struct {
}

func (s ServerCommand) Execute(args []string) error {

	mux, err := server.CreateMux()
	if err != nil {
		return err
	}
	return http.ListenAndServe(":3001", mux)
}
