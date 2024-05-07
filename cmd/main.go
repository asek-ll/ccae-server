package main

import (
	"errors"
	"os"

	"github.com/asek-ll/aecc-server/internal/cmd"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	ServerCmd   cmd.ServerCommand      `command:"server"`
	FillItems   cmd.FillItemsCommand   `command:"fill_items"`
	FillRecipes cmd.FillRecipesCommand `command:"fill_recipes"`
}

func main() {
	var options Options
	p := flags.NewParser(&options, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		return command.Execute(args)
	}
	if _, err := p.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
