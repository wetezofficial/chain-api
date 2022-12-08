/*
 * Created by Chengbin Du on 2022/4/25.
 */

package initapp

import (
	"net/http"
	"time"

	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/handler"
	"starnet/chain-api/pkg/proxy"
	"starnet/starnet/constant"
)

func initPolygonHandler(app *app.App) error {
	chain := constant.ChainPolygon

	httpBlackMethods := []string{
		"eth_getFilterChanges",
		"eth_getFilterLogs",
		"eth_newBlockFilter",
		"eth_newFilter",
		"eth_newPendingTransactionFilter",
		"eth_uninstallFilter",
		"eth_subscribe",
		"eth_unsubscribe",
	}

	var wsBlackMethods []string

	justWhiteMethods:=[]string{
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

	cacheableMethods := []string{
		"eth_blockNumber",
		"eth_getBlockByHash",
		"eth_getBlockByNumber",
		"eth_getTransactionByHash",
		"eth_getTransactionCount",
		"eth_getTransactionReceipt",
		"eth_getBlockTransactionCountByHash",
		"eth_getBlockTransactionCountByNumber",
		"eth_getTransactionByBlockNumberAndIndex",
		"bor_getAuthor",
		"bor_getCurrentValidators",
		"bor_getCurrentProposer",
		"bor_getRootHash",
		"eth_getRootHash",
		"eth_getSignersAtHash",
		"eth_getTransactionReceiptsByBlock",
		"eth_getTransactionByBlockHashAndIndex",
		"eth_getBalance",
		"eth_getCode",
		"eth_getStorageAt",
		"eth_accounts",
		"eth_getProof",
		"eth_getLogs",
		"eth_gasPrice",
		"eth_chainId",
		"net_version",
		"eth_getUncleByBlockNumberAndIndex",
		"eth_getUncleByBlockHashAndIndex",
		"eth_getUncleCountByBlockHash",
		"eth_getUncleCountByBlockNumber",
		"net_listening",
		"web3_clientVersion",
	}

	cfg := proxy.JsonRpcProxyConfig{
		HttpUpstream:     app.Config.Upstream.Polygon.Http,
		WsUpstream:       app.Config.Upstream.Polygon.Ws,
		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 5, // block time 2.3s https://www.blocknative.com/blog/monitor-polygon-mempool
		ChainID:          chain.ChainID,
		CacheableMethods: cacheableMethods,
	}

	p := proxy.NewJsonRpcProxy(app, cfg)

	h := handler.NewJsonRpcHandler(
		chain,
		httpBlackMethods,
		wsBlackMethods,
		justWhiteMethods,
		p,
		app,
	)

	app.PolygonHttpHandler = h
	app.PolygonWsHandler = h

	return nil
}
