/*
 * Created by Du, Chengbin on 2022/4/26.
 */

package proxy

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"starnet/chain-api/pkg/jsonrpc"
	"starnet/chain-api/pkg/utils"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func (p *JsonRpcProxy) fromRequestPath(tenderMintRequest jsonrpc.TenderMintRequest) (*request, error) {
	req := request{
		ID:                atomic.AddInt64(&p.requestID, 1),
		TenderMintRequest: &tenderMintRequest,
	}
	return &req, nil
}

func (p *JsonRpcProxy) TendermintProxy(ctx context.Context, logger *zap.Logger, tenderMintRequest jsonrpc.TenderMintRequest) ([]byte, error) {
	req, err := p.fromRequestPath(tenderMintRequest)
	if err != nil {
		return nil, err
	}
	req.ctx = ctx
	req.logger = logger

	var resp []byte
	resp, err = p.fromTendermintCache(req)
	if resp != nil || err != nil {
		return resp, err
	}

	return p.TendermintUpstream(req)
}

func (p *JsonRpcProxy) fromTendermintCache(req *request) ([]byte, error) {
	// step1. Try to get result from cache
	hash := md5.Sum([]byte(req.TenderMintRequest.Path + req.TenderMintRequest.URLQuery))
	cacheKey := fmt.Sprintf("tendermint:%d:%s:%s", p.cfg.ChainID, req.TenderMintRequest.Path, hex.EncodeToString(hash[:]))
	cacheable := utils.In(req.TenderMintRequest.Path, p.cfg.CacheableMethods)
	if cacheable {
		req.cacheKey = &cacheKey
		req.cacheFn = p.CacheFn

		res, err := p.rdb.Get(req.ctx, cacheKey).Bytes()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}

		// 获取到了内容，则直接返回
		if len(res) > 0 {
			req.logger.Debug("got resp from cache", zap.ByteString("cached resp", res))

			return res, nil
		}
	}

	return nil, nil
}

func (p *JsonRpcProxy) DoTendermintUpstreamCall(req *jsonrpc.TenderMintRequest) ([]byte, error) {
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

func (p *JsonRpcProxy) TendermintUpstream(req *request) ([]byte, error) {
	resp, err := p.DoTendermintUpstreamCall(req.TenderMintRequest)
	if err != nil {
		return nil, err
	}

	req.logger.Debug("new upstream response", zap.ByteString("resp", resp))

	upstreamResp := UpstreamJsonRpcResponse{}
	if err = json.Unmarshal(resp, &upstreamResp); err != nil {
		return nil, err
	}

	// step3. Cache if it is a valid result and cacheable
	if req.cacheKey != nil && upstreamResp.Result != nil {
		saveDate, err := json.Marshal(upstreamResp)
		if err != nil {
			return resp, nil
		}
		if err = p.CacheFn(req, saveDate); err != nil {
			req.logger.Error("failed to cache result", zap.Error(err))
		}
	}

	return resp, nil
}
