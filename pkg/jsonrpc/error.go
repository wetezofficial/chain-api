package jsonrpc

import (
	"encoding/json"
	"fmt"
)

var UnauthorizedErr = &JsonRpcErr{
	Code:    401,
	Message: "Unauthorized",
}

var TooManyRequestErr = &JsonRpcErr{
	Code:    429,
	Message: "Too many requests",
}

var ParseError = &JsonRpcErr{
	Code:    -32700,
	Message: "Parse error",
}

var FilterNotFoundError = &JsonRpcErr{
	Code:    -32000,
	Message: "filter not found",
}

var SubscriptionNotFoundError = &JsonRpcErr{
	Code:    -32000,
	Message: "subscription not found",
}

func NewInternalServerError(id interface{}) *JsonRpcErr {
	return &JsonRpcErr{
		ID:      id,
		Code:    -32000,
		Message: "Internal server error",
	}
}

func NewUnsupportedMethodError(id interface{}) *JsonRpcErr {
	return &JsonRpcErr{
		ID:      id,
		Code:    -32601,
		Message: "Unsupported method",
	}
}

type JsonRpcErr struct {
	ID      interface{}
	Code    int
	Message string
	Data    json.RawMessage
}

func (e *JsonRpcErr) MarshalJSON() ([]byte, error) {
	resp := map[string]interface{}{
		"id":      e.ID,
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    e.Code,
			"message": e.Message,
		},
	}
	if e.Data != nil {
		resp["error"].(map[string]interface{})["data"] = e.Data
	}

	return json.Marshal(resp)
}

func (e *JsonRpcErr) Error() string {
	return fmt.Sprintf("jsonrpc err with code: %d message: %s", e.Code, e.Message)
}
