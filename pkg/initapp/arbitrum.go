/*
 * Created by Chengbin Du on 2022/4/25.
 */

package initapp

import (
	"net/http"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/chain-api/pkg/jsonrpc"
	"starnet/chain-api/pkg/proxy"
	"starnet/starnet/constant"
	"time"
)

func initArbitrumHandler(app *app.App) error {
	chain := constant.ChainArbitrum
	httpSupportedMethods := []string{
		"eth_blockNumber",
		"eth_getBlockByHash",
		"eth_getBlockByNumber",
		"eth_getTransactionByHash",
		"eth_getTransactionCount",
		"eth_getTransactionReceipt",
		"eth_getBlockTransactionCountByHash",
		"eth_getBlockTransactionCountByNumber",
		"eth_getTransactionByBlockHashAndIndex",
		"eth_getTransactionByBlockNumberAndIndex",
		"eth_sendRawTransaction",
		"eth_getBalance",
		"eth_getCode",
		"eth_getStorageAt",
		"eth_accounts",
		"eth_call",
		"eth_getLogs",
		"eth_gasPrice",
		"eth_estimateGas",
		"net_version",
		"eth_chainId",
		"web3_clientVersion",
		"web3_sha3",
	}

	wsSupportedMethods := append(httpSupportedMethods, []string{
		"eth_getFilterChanges",
		"eth_getFilterLogs",
		"eth_newBlockFilter",
		"eth_newFilter",
		"eth_uninstallFilter",

		"eth_subscribe",
		"eth_unsubscribe",
	}...)

	cacheableMethods := []string{
		"eth_blockNumber",
		"eth_getBlockByHash",
		"eth_getBlockByNumber",
		"eth_getTransactionByHash",
		"eth_getTransactionCount",
		"eth_getTransactionReceipt",
		"eth_getBlockTransactionCountByHash",
		"eth_getBlockTransactionCountByNumber",
		"eth_getTransactionByBlockHashAndIndex",
		"eth_getTransactionByBlockNumberAndIndex",
		"eth_getBalance",
		"eth_getCode",
		"eth_getStorageAt",
		"eth_accounts",
		"eth_getLogs",
		"eth_gasPrice",
		"net_version",
		"eth_chainId",
		"web3_clientVersion",
	}

	cfg := proxy.JsonRpcProxyConfig{
		HttpUpstream:     app.Config.Upstream.Arbitrum.Http,
		WsUpstream:       app.Config.Upstream.Arbitrum.Ws,
		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 15, // L1 15 seconds L2 1 minutes  https://developer.offchainlabs.com/docs/time_in_arbitrum
		ChainID:          chain.ChainID,
		CacheableMethods: cacheableMethods,

		// filters
		FilterIDExtractor: jsonrpc.EthFilterIDExtractor,
		CreateFilterMethods: []string{
			"eth_newBlockFilter",
			"eth_newFilter",
		},
		FilterMethods: []string{
			"eth_getFilterChanges",
			"eth_getFilterLogs",
		},
		UninstallFilterMethods: []string{
			"eth_uninstallFilter",
		},

		// subscriptions
		SubscriptionIDExtractor: jsonrpc.EthSubscriptionIDExtractor,
		SubscribeMethods: []string{
			"eth_subscribe",
		},
		UnsubscribeMethods: []string{
			"eth_unsubscribe",
		},
	}

	p := proxy.NewJsonRpcProxy(app, cfg)

	h := handler.NewJsonRpcHandler(
		chain,
		httpSupportedMethods,
		wsSupportedMethods,
		cfg.SubscribeMethods,
		p,
		app,
	)

	app.ArbitrumHttpHandler = h
	app.ArbitrumWsHandler = h

	return nil
}
