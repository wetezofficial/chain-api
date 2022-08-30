package initapp

var (
	tendermintHttpBlackMethods = []string{
		"dial_seeds",
		"dial_peers",
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

		// append 20220830
		"commit",
		"net_info",
		"blockchain",
		"validators",
		"dump_consensus_state",
		"consensus_state",
		"consensus_params",

		// append 20220830 && before in tendermintHttpBlackMethods
		"genesis",         // rpc return use genesis_chunked
		"genesis_chunked", // data too big
		"tx_search",       // fixme: "error converting http params to arguments: invalid character 'x' in literal true (expecting 'r')"
		"abci_query",      // fixme: error converting http params or panic message
	}
)
