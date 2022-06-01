/*
 * Created by Chengbin Du on 2022/4/26.
 */

package jsonrpc

import (
	"encoding/json"
	"strconv"
)

func SolanaSubscriptionIDExtractor(req *JsonRpcRequest) (string, error) {
	var id []int
	err := json.Unmarshal(req.Params, &id)
	if err != nil {
		return "", err
	}
	if len(id) != 1 {
		return "", ParseError
	}
	return strconv.Itoa(id[0]), nil
}

func EthSubscriptionIDExtractor(req *JsonRpcRequest) (string, error) {
	var id []string
	err := json.Unmarshal(req.Params, &id)
	if err != nil {
		return "", err
	}
	if len(id) != 1 {
		return "", ParseError
	}
	return id[0], nil
}

func EthFilterIDExtractor(req *JsonRpcRequest) (string, error) {
	var id []string
	err := json.Unmarshal(req.Params, &id)
	if err != nil {
		return "", err
	}
	if len(id) != 1 {
		return "", ParseError
	}
	return id[0], nil
}
