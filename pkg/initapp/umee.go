package initapp

import (
	"net/http"
	"time"

	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/chain-api/pkg/proxy"
	"starnet/starnet/constant"
)

func initUmeeHandler(app *app.App) error {
	chain := constant.ChainUmee
	chainUpstreamCfg := app.Config.Upstream.Umee

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
		[]string{},
		tendermintWsBlackMethods,
		justWhiteMethods,
		p,
		app,
	)

	app.UmeeHttpHandler = h
	app.UmeeWsHandler = h

	return nil
}
