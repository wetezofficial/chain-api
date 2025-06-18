package config

import (
	"fmt"
	"testing"
)

func TestLoadRPCConfig(t *testing.T) {
	rpcConfig, err := LoadRPCConfig(`
[lisk]
http = "https://api.lisk.com"
ws = "wss://api.lisk.com"
backup_http = "https://api.lisk.com"
backup_ws = "wss://api.lisk.com"
max_behind_blocks = 10
block_number_method = "getBlockNumber"
block_number_result_extractor = "jq"
block_number_result_expression = ".result"
	`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", rpcConfig)
}
