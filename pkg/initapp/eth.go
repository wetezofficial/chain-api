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

func initEthHandler(app *app.App) error {
	chain := constant.ChainETH
	// 不支持的命令有:
	// eth_coinbase
	// eth_mining
	// eth_hashrate
	// eth_getWork
	// eth_submitWork
	// eth_sign
	// eth_submitHashrate
	// eth_syncing
	// eth_sendTransaction
	// eth_signTransaction
	// eth_accounts

	httpSupportedMethods := []string{
		"eth_getBlockByHash",
		"eth_getBlockByNumber",
		"eth_getBlockTransactionCountByHash",
		"eth_getBlockTransactionCountByNumber",
		"eth_getUncleCountByBlockHash",
		"eth_getUncleCountByBlockNumber",
		"eth_protocolVersion",
		"eth_chainId",
		"eth_blockNumber",
		"eth_call",
		"eth_estimateGas",
		"eth_gasPrice",
		"eth_feeHistory",
		"eth_getLogs",
		"eth_getBalance",
		"eth_getStorageAt",
		"eth_getTransactionCount",
		"eth_getCode",
		"eth_sendRawTransaction",
		"eth_getTransactionByHash",
		"eth_getTransactionByBlockHashAndIndex",
		"eth_getTransactionByBlockNumberAndIndex",
		"eth_getTransactionReceipt",
		"net_version",
		"net_listening",
		"web3_clientVersion",
		"web3_sha3",
	}

	// WS 方式支持 filter 及 subscription
	wsSupportedMethods := append(httpSupportedMethods, []string{
		"eth_newFilter",
		"eth_newBlockFilter",
		"eth_newPendingTransactionFilter",
		"eth_uninstallFilter",
		"eth_getFilterChanges",
		"eth_getFilterLogs",

		"eth_subscribe",
		"eth_unsubscribe",
	}...)

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
		HttpUpstream:     app.Config.Upstream.Eth.Http,
		WsUpstream:       app.Config.Upstream.Eth.Ws,
		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 12,
		ChainID:          chain.ChainID,
		CacheableMethods: cacheableMethods,

		// filters
		FilterIDExtractor: jsonrpc.EthFilterIDExtractor,
		CreateFilterMethods: []string{
			"eth_newBlockFilter",
			"eth_newFilter",
			"eth_newPendingTransactionFilter",
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

	app.EthHttpHandler = h
	app.EthWsHandler = h

	return nil
}
