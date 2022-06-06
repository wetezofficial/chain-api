package jsonrpc

import (
	"bytes"
	"encoding/json"
)

// https://www.jsonrpc.org/specification#compatibility

type JsonRpcRequest struct {
	batchCall  []JsonRpcSingleRequest
	singleCall *JsonRpcSingleRequest
}

func (r *JsonRpcRequest) IsBatchCall() bool {
	return r.batchCall != nil
}

func (r *JsonRpcRequest) Cost() int {
	if r.singleCall != nil {
		return 1
	}

	n := len(r.batchCall)
	if n == 0 {
		return 1
	}
	return n
}

func (r *JsonRpcRequest) GetBatchCall() []JsonRpcSingleRequest {
	return r.batchCall
}

func (r *JsonRpcRequest) GetSingleCall() *JsonRpcSingleRequest {
	return r.singleCall
}

func (r *JsonRpcRequest) MarshalJSON() ([]byte, error) {
	if r.batchCall != nil {
		return json.Marshal(r.batchCall)
	}
	return json.Marshal(r.singleCall)
}

func (r *JsonRpcRequest) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if data[0] == '[' && data[len(data)-1] == ']' {
		// Batch call
		return json.Unmarshal(data, &r.batchCall)
	}
	// single request
	return json.Unmarshal(data, &r.singleCall)
}

type JsonRpcSingleRequest struct {
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
