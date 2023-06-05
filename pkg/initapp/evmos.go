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

	cfg := proxy.JsonRpcProxyConfig{
		HttpUpstream:     chainUpstreamCfg.Http,
		WsUpstream:       chainUpstreamCfg.Ws,
		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 4,  // block time 4.26s https://escan.live/
		ChainID:          chain.ChainID,
		CacheableMethods: tendermintCacheableMethods,
	}

	p := proxy.NewJsonRpcProxy(app, cfg)

	var justWhiteMethods []string

	h := handler.NewJsonRpcHandler(
		chain,
		tendermintHttpBlackMethods,
		[]string{},
		tendermintWsBlackMethods,
		justWhiteMethods,
		p,
		app,
	)

	app.EvmosHttpHandler = h
	app.EvmosWsHandler = h

	return nil
}
