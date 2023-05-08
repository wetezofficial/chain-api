/*
 * Created by Chengbin Du on 2022/4/25.
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

func initEthHandler(app *app.App) error {
	chain := constant.ChainETH

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

	erigonMethods := []string{
		"eth_getLogs",
		"eth_getBlockReceipts",
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

	justWhiteMethods := []string{
		"trace_call",
		"trace_block",
		"trace_get",
		"trace_filter",
		"trace_transaction",
		"trace_rawTransaction",
		"trace_replayBlockTransactions",
		"trace_replayTransaction",

		"debug_traceCall",
		"debug_traceTransaction",
		"debug_traceBlockByNumber",
		"debug_traceBlockByHash",
	}

	cfg := proxy.JsonRpcProxyConfig{
		HttpUpstream: app.Config.Upstream.Eth.Http,
		WsUpstream:   app.Config.Upstream.Eth.Ws,

		HttpErigonStream: app.Config.Upstream.Eth.Erigon.Http,
		WsErigonUpstream: app.Config.Upstream.Eth.Erigon.Ws,

		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 12,
		ChainID:          chain.ChainID,
		CacheableMethods: cacheableMethods,
	}

	p := proxy.NewJsonRpcProxy(app, cfg)

	h := handler.NewJsonRpcHandler(
		chain,
		httpBlackMethods,
		erigonMethods,
		wsBlackMethods,
		justWhiteMethods,
		p,
		app,
	)

	app.EthHttpHandler = h
	app.EthWsHandler = h

	return nil
}
