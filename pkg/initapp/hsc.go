/*
 * Created by Chengbin Du on 2022/6/9.
 */

package initapp

import (
	"net/http"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/chain-api/pkg/proxy"
	"starnet/starnet/constant"
	"time"
)

func initHscHandler(app *app.App) error {
	chain := constant.ChainHSC
	chainUpstreamCfg := app.Config.Upstream.Hsc

	httpBlackMethods := []string{
		"eth_newFilter",
		"eth_newBlockFilter",
		"eth_newPendingTransactionFilter",
		"eth_uninstallFilter",
		"eth_getFilterChanges",
		"eth_getFilterLogs",
		"eth_subscribe",
		"eth_unsubscribe",
	}
	var wsBlackMethods []string

	cacheableMethods := []string{
		"eth_getBlockByHash",
		"eth_getBlockByNumber",
		"eth_getBlockTransactionCountByHash",
		"eth_getBlockTransactionCountByNumber",
		"eth_getUncleCountByBlockHash",
		"eth_getUncleCountByBlockNumber",
		"eth_protocolVersion",
		"eth_chainId",
		"eth_blockNumber",
		"eth_gasPrice",
		"eth_feeHistory",
		"eth_getBalance",
		"eth_getStorageAt",
		"eth_getTransactionCount",
		"eth_getCode",
		"eth_getTransactionByHash",
		"eth_getTransactionByBlockHashAndIndex",
		"eth_getTransactionByBlockNumberAndIndex",
		"eth_getTransactionReceipt",
		"net_version",
		"net_listening",
		"web3_clientVersion",
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

	app.HscHttpHandler = h
	app.HscWsHandler = h

	return nil
}
