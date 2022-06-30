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

	var httpBlockMethods []string
	var wsBlockMethods []string

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
	}

	p := proxy.NewJsonRpcProxy(app, cfg)

	h := handler.NewJsonRpcHandler(
		chain,
		httpBlockMethods,
		wsBlockMethods,
		p,
		app,
	)

	app.EthHttpHandler = h
	app.EthWsHandler = h

	return nil
}
