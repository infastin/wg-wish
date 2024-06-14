package app

import (
	"github.com/alecthomas/kong"
)

type CLI struct {
	Config string `optional:"" short:"c" type:"existingfile" placeholder:"PATH" help:"Path to the config file."`
}

func NewCLI(args []string) (cli CLI, err error) {
	k, err := kong.New(&cli)
	if err != nil {
		return CLI{}, err
	}

	_, err = k.Parse(args[1:])
	if err != nil {
		return CLI{}, err
	}

	return cli, nil
}
