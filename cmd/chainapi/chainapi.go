package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"starnet/chain-api/pkg/initapp"
	"starnet/starnet"
)

func main() {
	os.Setenv("TZ", "UTC")

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
		},
		Action: func(c *cli.Context) error {
			fmt.Println("Starnet Chain API service")
			fmt.Println("Starnet Version", starnet.Version)

			configFile := c.String("config")
			app := initapp.NewApp(configFile)
			app.Start()

			return nil
		},
	}

	cliApp.Run(os.Args)
}
