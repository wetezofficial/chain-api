package initapp

var (
	tendermintHttpBlackMethods = []string{
		"genesis",         // rpc return use genesis_chunked
		"genesis_chunked", // data too big
		"tx_search",       // fixme: "error converting http params to arguments: invalid character 'x' in literal true (expecting 'r')"
		"abci_query",      // fixme: error converting http params or panic message
	}

	tendermintWsBlackMethods []string

	tendermintCacheableMethods = []string{
		"abci_info",
		"block",
		"block_by_hash",
		"block_results",
		"block_search",
		"blockchain",
		"health",
		"status",
		"commit",
		"validators",
		"genesis",
		"unconfirmed_txs",
		"num_unconfirmed_txs",
		"tx",
	}
)
