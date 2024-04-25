package cmd

import (
	"fmt"

	"github.com/jessevdk/go-flags"
)

var _ flags.Commander = ServerCommand{}

type ServerCommand struct {
}

func (s ServerCommand) Execute(args []string) error {
	fmt.Println("Hello")
	return nil
}
