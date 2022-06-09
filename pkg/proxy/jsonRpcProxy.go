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
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"starnet/chain-api/pkg/app"
	"starnet/chain-api/pkg/jsonrpc"
	"starnet/chain-api/pkg/utils"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type UpstreamJsonRpcResponse struct {
	ID uint64 `json:"id"`
	// JsonRpcVersion A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
	JsonRpcVersion string          `json:"jsonrpc"`
	Error          json.RawMessage `json:"error,omitempty"`
	Result         json.RawMessage `json:"result,omitempty"`
}

type request struct {
	*jsonrpc.JsonRpcRequest
	ID uint64 `json:"id"` // overwrite id while sending to upstream

	cacheKey *string
	cacheFn  func(request *request, result []byte) error
	ctx      context.Context
	logger   *zap.Logger
}

type RespData struct {
	Data         []byte
	Subscription bool
}

type JsonRpcProxyConfig struct {
	HttpUpstream string
	WsUpstream   string

	HttpClient       *http.Client
	CacheTime        time.Duration
	ChainID          uint8
	CacheableMethods []string
}

type JsonRpcProxy struct {
	rdb        redis.UniversalClient
	httpClient *http.Client
	cfg        *JsonRpcProxyConfig
	requestID  uint64
}

func NewJsonRpcProxy(app *app.App, cfg JsonRpcProxyConfig) *JsonRpcProxy {
	p := &JsonRpcProxy{
		rdb:        app.Rdb,
		httpClient: cfg.HttpClient,
		cfg:        &cfg,
	}

	return p
}

func (p *JsonRpcProxy) fromRequest(rawreq *jsonrpc.JsonRpcRequest) (*request, error) {
	req := request{
		JsonRpcRequest: rawreq,
		ID:             atomic.AddUint64(&p.requestID, 1),
	}
	return &req, nil
}

func (p *JsonRpcProxy) HttpProxy(ctx context.Context, logger *zap.Logger, rawreq *jsonrpc.JsonRpcRequest) ([]byte, error) {
	if rawreq.IsBatchCall() {
		return p.DoHttpUpstreamCall(rawreq)
	}

	req, err := p.fromRequest(rawreq)
	if err != nil {
		return nil, err
	}
	req.ctx = ctx
	req.logger = logger

	var resp []byte
	resp, err = p.fromCache(req)
	if resp != nil || err != nil {
		return resp, err
	}

	return p.HttpUpstream(req)
}

// fromCache 尝试从缓存中读取数据
func (p *JsonRpcProxy) fromCache(req *request) ([]byte, error) {
	singleReq := req.GetSingleCall()

	// step1. Try to get result from cache
	hash := md5.Sum(singleReq.Params)
	cacheKey := fmt.Sprintf("rpc:%d:%s:%s", p.cfg.ChainID, singleReq.Method, hex.EncodeToString(hash[:]))
	cacheable := utils.In(singleReq.Method, p.cfg.CacheableMethods)
	if cacheable {
		req.cacheKey = &cacheKey
		req.cacheFn = p.CacheFn

		res, err := p.rdb.Get(req.ctx, cacheKey).Bytes()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}

		// 获取到了内容，则直接返回
		if len(res) > 0 {
			resp := jsonrpc.JsonRpcResponse{
				ID:             singleReq.ID,
				JsonRpcVersion: singleReq.JsonRpcVersion,
				Result:         res,
			}
			data, err := json.Marshal(resp)
			if err != nil {
				return nil, err
			}

			req.logger.Debug("got resp from cache", zap.ByteString("cached resp", data))

			return data, nil
		}
	}

	return nil, nil
}

func (p *JsonRpcProxy) CacheFn(req *request, result []byte) error {
	return p.rdb.Set(context.TODO(), *req.cacheKey, result, p.cfg.CacheTime).Err()
}

func (p *JsonRpcProxy) DoHttpUpstreamCall(req *jsonrpc.JsonRpcRequest) ([]byte, error) {
	rawreq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	res, err := p.httpClient.Post(p.cfg.HttpUpstream, "application/json", strings.NewReader(string(rawreq)))
	if err != nil {
		return nil, err
	}

	buff := bytes.Buffer{}
	_, err = buff.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (p *JsonRpcProxy) HttpUpstream(req *request) ([]byte, error) {
	resp, err := p.DoHttpUpstreamCall(req.JsonRpcRequest)
	req.logger.Debug("new upstream response", zap.ByteString("resp", resp))

	upstreamResp := UpstreamJsonRpcResponse{}
	if err = json.Unmarshal(resp, &upstreamResp); err != nil {
		return nil, err
	}

	// step3. Cache if it is a valid result and cacheable
	if req.cacheKey != nil && upstreamResp.Result != nil {
		if err = p.CacheFn(req, upstreamResp.Result); err != nil {
			req.logger.Error("failed to cache result", zap.Error(err))
		}
	}

	return resp, nil
}

func (p *JsonRpcProxy) NewUpstreamWS(client *Client, logger *zap.Logger) (*UpstreamWebSocket, error) {
	upstream, _, err := websocket.DefaultDialer.Dial(p.cfg.WsUpstream, nil)
	if err != nil {
		return nil, err
	}

	u := &UpstreamWebSocket{
		conn:     upstream,
		client:   client,
		logger:   logger,
		proxy:    p,
		mutex:    new(sync.Mutex),
		requests: make(map[uint64]*request),
	}
	go u.run()
	return u, nil
}
