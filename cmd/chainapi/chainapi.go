package main

import (
	"fmt"
	"os"

	"starnet/chain-api/pkg/initapp"
	"starnet/starnet"

	"github.com/urfave/cli/v2"
)

func main() {
	err := os.Setenv("TZ", "UTC")
	if err != nil {
		panic(err)
	}

	cliApp := &cli.App{
		Name:  "chainapi",
		Usage: "Chain API service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "",
				Aliases: []string{
					"C",
				},
				Usage:    "specify config",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "rpc_config",
				Value:    "",
				Usage:    "specify RPC config file",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println("Starnet Chain API service")
			fmt.Println("Starnet Version", starnet.Version)

			configFile := c.String("config")
			app := initapp.NewApp(configFile, c.String("rpc_config"))
			app.Start()

			return nil
		},
	}

	_ = cliApp.Run(os.Args)
}
