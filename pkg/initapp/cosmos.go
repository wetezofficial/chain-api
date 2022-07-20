package initapp

import (
	"net/http"
	"time"

	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/chain-api/pkg/proxy"
	"starnet/starnet/constant"
)

func initCosmosHandler(app *app.App) error {
	chain := constant.ChainCosmos
	chainUpstreamCfg := app.Config.Upstream.Cosmos

	httpBlackMethods := []string{}
	var wsBlackMethods []string

	// TODO:
	cacheableMethods := []string{
		"abci_info",
		"block",
		"block_by_hash",
		"block_results",
		"block_search",
		"blockchain",
	}

	cfg := proxy.JsonRpcProxyConfig{
		HttpUpstream:     chainUpstreamCfg.Http,
		WsUpstream:       chainUpstreamCfg.Ws,
		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 6, // block time 3s
		ChainID:          chain.ChainID,
		CacheableMethods: cacheableMethods,
	}

	p := proxy.NewJsonRpcProxy(app, cfg)

	h := handler.NewJsonRpcHandler(
		chain,
		httpBlackMethods,
		wsBlackMethods,
		p,
		app,
	)

	app.CosmosHttpHandler = h
	app.CosmosWsHandler = h

	return nil
}
