package handler

import (
	"fmt"
	"testing"

	"github.com/itchyny/gojq"
)

func TestGetBlockNumberFromEvmHttp(t *testing.T) {
	jqQuery, err := gojq.Parse(".result")
	if err != nil {
		t.Fatal(err)
	}
	got, err := getBlockNumberFromEvmHttp("https://ethereum-rpc.publicnode.com", `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`, jqQuery)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("got", got)
}

func TestGetBlockNumberFromWs(t *testing.T) {
	jqQuery, err := gojq.Parse(".result")
	if err != nil {
		t.Fatal(err)
	}
	got, err := getBlockNumberFromEvmWs("wss://ethereum-rpc.publicnode.com", `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`, jqQuery)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("got", got)
}
