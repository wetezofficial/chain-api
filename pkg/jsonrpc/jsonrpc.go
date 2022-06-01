package jsonrpc

import (
	"encoding/json"
)

// https://www.jsonrpc.org/specification#compatibility

type JsonRpcRequest struct {
	ID *interface{} `json:"id"`
	// JsonRpcVersion A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JsonRpcVersion string          `json:"jsonrpc"`
	Method         string          `json:"method"`
	Params         json.RawMessage `json:"params,omitempty"`
}

type JsonRpcResponse struct {
	ID *interface{} `json:"id"`
	// JsonRpcVersion A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JsonRpcVersion string          `json:"jsonrpc"`
	Error          json.RawMessage `json:"error,omitempty"`
	Result         json.RawMessage `json:"result,omitempty"`
}

// SubscriptionNotification
// https://geth.ethereum.org/docs/rpc/pubsub
// 不支持 syncing method，所以无需对 syncing 返回的特殊数据结构进行处理
type SubscriptionNotification struct {
	// JsonRpcVersion A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JsonRpcVersion string `json:"jsonrpc"`
	Method         string `json:"method"`
	Params         struct {
		Subscription string          `json:"subscription"`
		Result       json.RawMessage `json:"result"`
	} `json:"params"`
}
