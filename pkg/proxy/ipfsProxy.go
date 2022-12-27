// Package proxy TODO: .
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"starnet/chain-api/pkg/jsonrpc"
)

func (p *JsonRpcProxy) IPFSProxy(ctx context.Context, logger *zap.Logger) ([]byte, error) {
	return p.IPFSUpstream(nil)
}

func (p *JsonRpcProxy) IPFSUpstream(req *request) ([]byte, error) {
	resp, err := p.DoIPFSUpstreamCall(req.TenderMintRequest)
	if err != nil {
		return nil, err
	}

	req.logger.Debug("new upstream response", zap.ByteString("resp", resp))

	upstreamResp := UpstreamJsonRpcResponse{}
	if err = json.Unmarshal(resp, &upstreamResp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *JsonRpcProxy) DoIPFSUpstreamCall(req *jsonrpc.TenderMintRequest) ([]byte, error) {
	res, err := p.httpClient.Get(p.cfg.HttpUpstream + "/" + req.Path + req.URLQuery)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf(res.Status)
	}

	buff := bytes.Buffer{}
	_, err = buff.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
