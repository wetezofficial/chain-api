package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
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
			&cli.BoolFlag{
				Name:     "eth_blockNumber",
				Value:    false,
				Usage:    "send eth_blockNumber message to upstream websocket",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			ethBlockNumber := c.Bool("eth_blockNumber")

			logger, _ := zap.NewProduction()
			defer logger.Sync() // flushes buffer, if any

			// Connect to the upstream WebSocket server
			reqHeader := http.Header{}
			reqHeader.Set("X-Forwarded-For", "127.0.0.1")
			reqHeader.Set("X-Real-Ip", "127.0.0.1")
			url := c.Args().Get(0)
			upstream, _, err := websocket.DefaultDialer.Dial(url, reqHeader)
			if err != nil {
				logger.Error("failed to dial upstream websocket", zap.Error(err), zap.String("url", url))
				return err
			}
			defer upstream.Close()
			fmt.Println("connected to websocket")

			if ethBlockNumber {
				msg := `{"id":0,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}`
				fmt.Printf("> %s\n", msg)
				err = upstream.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					logger.Error("failed to write message to upstream websocket", zap.Error(err))
					return err
				}

				_, message, err := upstream.ReadMessage()
				if err != nil {
					logger.Error("failed to read message from upstream websocket", zap.Error(err))
					return err
				}
				fmt.Printf("< %s\n", string(message))
				upstream.Close()
				return nil
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				fmt.Print("> ")
				for {
					select {
					case <-ctx.Done():
						fmt.Println("received signal, exiting")
						return
					default:
						reader := bufio.NewReader(os.Stdin)
						message, _ := reader.ReadString('\n')
						time.Sleep(5 * time.Second)
						upstream.WriteMessage(websocket.TextMessage, []byte(message))
					}
				}
			}()

			go func() {
				for {

					select {
					case <-ctx.Done():
						fmt.Println("received signal, exiting")
						return
					default:
						_, message, err := upstream.ReadMessage()
						if err != nil {
							logger.Error("failed to read message from upstream websocket", zap.Error(err))
							return
						}
						fmt.Printf("< %s\n> ", string(message))
					}

				}
			}()
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			<-ch
			fmt.Println("received signal, exiting")
			upstream.Close()

			return nil
		},
	}

	_ = cliApp.Run(os.Args)
}
