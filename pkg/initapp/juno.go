package initapp

import (
	"net/http"
	"time"

	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/chain-api/pkg/proxy"
	"starnet/starnet/constant"
)

func initJunoHandler(app *app.App) error {
	chain := constant.ChainJuno
	chainUpstreamCfg := app.Config.Upstream.Juno

	cfg := proxy.JsonRpcProxyConfig{
		HttpUpstream:     chainUpstreamCfg.Http,
		WsUpstream:       chainUpstreamCfg.Ws,
		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 6, // block time 3s
		ChainID:          chain.ChainID,
		CacheableMethods: tendermintCacheableMethods,
	}

	p := proxy.NewJsonRpcProxy(app, cfg)

	var justWhiteMethods []string

	h := handler.NewJsonRpcHandler(
		chain,
		tendermintHttpBlackMethods,
		tendermintWsBlackMethods,
		justWhiteMethods,
		p,
		app,
	)

	app.JunoHttpHandler = h
	app.JunoWsHandler = h

	return nil
}
