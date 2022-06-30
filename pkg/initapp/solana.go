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

func initSolanaHandler(app *app.App) error {
	chain := constant.ChainSolana

	var httpBlackMethods []string
	var wsBlackMethods []string

	cacheableMethods := []string{
		"getAccountInfo",
		"getBalance",
		"getBlock",
		"getBlockHeight",
		"getBlockProduction",
		"getBlockCommitment",
		"getBlocks",
		"getBlocksWithLimit",
		"getBlockTime",
		"getClusterNodes",
		"getEpochInfo",
		"getEpochSchedule",
		"getFeeForMessage",
		"getFirstAvailableBlock",
		"getGenesisHash",
		"getHealth",
		"getHighestSnapshotSlot",
		"getIdentity",
		"getInflationGovernor",
		"getInflationRate",
		"getInflationReward",
		"getLargestAccounts",
		"getLatestBlockhash",
		"getLeaderSchedule",
		"getMaxRetransmitSlot",
		"getMaxShredInsertSlot",
		"getMinimumBalanceForRentExemption",
		"getMultipleAccounts",
		"getProgramAccounts",
		"getRecentPerformanceSamples",
		"getSignaturesForAddress",
		"getSignatureStatuses",
		"getSlot",
		"getSlotLeader",
		"getSlotLeaders",
		"getStakeActivation",
		"getSupply",
		"getTokenAccountBalance",
		"getTokenAccountsByDelegate",
		"getTokenAccountsByOwner",
		"getTokenLargestAccounts",
		"getTokenSupply",
		"getTransaction",
		"getTransactionCount",
		"getVersion",
		"getVoteAccounts",
		"isBlockhashValid",
		"minimumLedgerSlot",

		// Deprecated methods
		"getConfirmedBlock",
		"getConfirmedBlocks",
		"getConfirmedBlocksWithLimit",
		"getConfirmedSignaturesForAddress2",
		"getConfirmedTransaction",
		"getFeeCalculatorForBlockhash",
		"getFeeRateGovernor",
		"getFees",
		"getRecentBlockhash",
		"getSnapshotSlot",
	}

	cfg := proxy.JsonRpcProxyConfig{
		HttpUpstream:     app.Config.Upstream.Solana.Http,
		WsUpstream:       app.Config.Upstream.Solana.Ws,
		HttpClient:       http.DefaultClient,
		CacheTime:        time.Second * 4, // block time 400ms https://www.finextra.com/blogposting/21693/introduction-to-the-solana-blockchain
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

	app.SolanaHttpHandler = h
	app.SolanaWsHandler = h

	return nil
}
