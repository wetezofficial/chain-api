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

	httpBlackMethods := []string{
		"genesis", // rpc return use genesis_chunked
		"genesis_chunked", // data too big
		"tx_search", // fixme: "error converting http params to arguments: invalid character 'x' in literal true (expecting 'r')"
		"abci_query", // fixme: error converting http params or panic message
	}

	var wsBlackMethods []string

	cacheableMethods := []string{
		"abci_info",
		"block",
		"block_by_hash",
		"block_results",
		"block_search",
		"blockchain",
		"health",
		"status",
		"commit",
		"validators",
		"genesis",
		"unconfirmed_txs",
		"num_unconfirmed_txs",
		"tx",
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
