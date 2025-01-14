package main

import (
	"errors"
	"os"

	"github.com/asek-ll/aecc-server/internal/cmd"
	"github.com/jessevdk/go-flags"
)

type Params struct {
	Config string `short:"c" long:"config" description:"Config file location" default:"config.json"`
	DB     string `short:"d" long:"database" description:"Database file location" default:"data.db"`
}

type Options struct {
	Params

	ServerCmd         cmd.ServerCommand            `command:"server"`
	FillItems         cmd.FillItemsCommand         `command:"fill_items"`
	FillRecipes       cmd.FillRecipesCommand       `command:"fill_recipes"`
	FillInGameRecipes cmd.FillInGameRecipesCommand `command:"fill_from_game"`
}

func main() {
	var options Options
	p := flags.NewParser(&options, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		if c, ok := command.(*cmd.ServerCommand); ok {
			c.SetParams(&cmd.ServerCommandParameters{
				Config: options.Config,
				DB:     options.DB,
			})
		}
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
