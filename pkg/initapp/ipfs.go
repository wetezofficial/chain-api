package initapp

import (
	"context"
	"fmt"
	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	ma "github.com/multiformats/go-multiaddr"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/starnet/constant"
)

func initIPFSClient(app *app.App) error {
	chain := constant.ChainIPFS
	app.IPFSHandler = handler.NewIPFSHandler(chain, app)
	return nil
}

func NewIPFSClient() error {
	ctx := context.Background()
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 9094))
	if err != nil {
		return err
	}

	cfg := &client.Config{
		APIAddr:           addr,
		DisableKeepAlives: true,
	}
	c, err := client.NewDefaultClient(cfg)
	if err != nil {
		return err
	}
	out := make(chan api.AddedOutput)
	go func() {
		err = c.Add(ctx, []string{"filePath"}, api.AddParams{}, out)
		if err != nil {
			return
		}
	}()
	result := <-out
	fmt.Println(result.Cid)

	return nil
}
