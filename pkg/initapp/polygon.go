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

func initPolygonHandler(app *app.App) error {
	chain := constant.ChainPolygon

	var httpBlockMethods []string
	var wsBlockMethods []string

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
		httpBlockMethods,
		wsBlockMethods,
		p,
		app,
	)

	app.PolygonHttpHandler = h
	app.PolygonWsHandler = h

	return nil
}
