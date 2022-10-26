package initapp

import (
	"net/http"
	"time"

	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/chain-api/pkg/proxy"
	"starnet/starnet/constant"
)

func initIRISnetHandler(app *app.App) error {
	chain := constant.ChainIRISnet
	chainUpstreamCfg := app.Config.Upstream.IRISnet

	cfg := proxy.JsonRpcProxyConfig{
		HttpUpstream:     chainUpstreamCfg.Http,
		WsUpstream:       chainUpstreamCfg.Ws,
		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 6, // block time 3s
		ChainID:          chain.ChainID,
		CacheableMethods: tendermintCacheableMethods,
	}

	p := proxy.NewJsonRpcProxy(app, cfg)

	h := handler.NewJsonRpcHandler(
		chain,
		tendermintHttpBlackMethods,
		tendermintWsBlackMethods,
		p,
		app,
	)

	app.IRISnetHttpHandler = h
	app.IRISnetWsHandler = h

	return nil
}