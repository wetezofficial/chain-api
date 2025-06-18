package utils

import (
	"testing"

	"github.com/itchyny/gojq"
)

func TestJqQueryFirst(t *testing.T) {
	query, err := gojq.Parse(".result")
	if err != nil {
		t.Fatal(err)
	}
	result, err := JqQueryFirst([]byte(`{"jsonrpc":"2.0","id":0,"result":"0x61280"}`), query)
	if err != nil {
		t.Fatal(err)
	}
	if result != "0x61280" {
		t.Fatal("expected 0x61280, got ", result)
	}
	result, err = JqQueryFirst([]byte(`{"jsonrpc":"2.0","id":0,"result":123}`), query)
	if err != nil {
		t.Fatal(err)
	}
	if result != "123" {
		t.Fatal("expected 123, got ", result)
	}
}
