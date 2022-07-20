package initapp

import (
	"net/http"
	"time"

	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/chain-api/pkg/proxy"
	"starnet/starnet/constant"
)

func initEvmosHandler(app *app.App) error {
	chain := constant.ChainEvmos
	chainUpstreamCfg := app.Config.Upstream.Evmos

	httpBlackMethods := []string{}
	var wsBlackMethods []string

	// TODO:
	cacheableMethods := []string{}

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

	app.EvmosHttpHandler = h
	app.EvmosWsHandler = h

	return nil
}
