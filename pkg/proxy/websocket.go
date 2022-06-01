/*
 * Created by Du, Chengbin on 2022/4/27.
 */

package proxy

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"starnet/chain-api/pkg/jsonrpc"
)

type Client struct {
	conn *websocket.Conn
	send chan RespData
}

func NewClient(conn *websocket.Conn, send chan RespData) *Client {
	return &Client{conn: conn, send: send}
}

type UpstreamWebSocket struct {
	conn   *websocket.Conn
	client *Client
	proxy  *JsonRpcProxy

	requests map[uint64]*request
}

func (u *UpstreamWebSocket) Close() error {
	return u.conn.Close()
}

func (u *UpstreamWebSocket) Send(ctx context.Context, logger *zap.Logger, rawreq *jsonrpc.JsonRpcRequest) error {
	p := u.proxy
	req, err := p.fromRequest(rawreq)
	if err != nil {
		return err
	}
	req.ctx = ctx
	req.logger = logger

	var resp []byte
	resp, err = p.fromCache(req)
	if err != nil {
		return err
	}
	if resp != nil {
		u.client.send <- RespData{Data: resp}
		return nil
	}

	u.requests[req.ID] = req
	return u.conn.WriteJSON(req)
}

func (u *UpstreamWebSocket) run() {
	defer u.conn.Close()
	defer u.client.conn.Close()

	p := u.proxy
	ws := u.conn
	for {
		_, rawresp, err := ws.ReadMessage()
		if err != nil {
			//logger.Info("ws.ReadMessage", zap.Error(err))
			return
		}

		//req.logger.Debug("new upstream response", zap.ByteString("resp", rawreq))
		upstreamResp := UpstreamJsonRpcResponse{}
		if err = json.Unmarshal(rawresp, &upstreamResp); err != nil {
			//req.logger.Error("failed to cache result", zap.Error(err))
			return
		}

		// 订阅的通知是没有 id 字段的
		if upstreamResp.ID == 0 {
			// 直接把内容写入到客户端
			u.client.send <- RespData{Data: rawresp, Subscription: true}
			continue
		}

		req, ok := u.requests[upstreamResp.ID]
		if !ok {
			//req.logger.Error("could not found request", zap.Error(err))
			return
		}
		delete(u.requests, upstreamResp.ID)

		// step3. Cache if it is a valid result and cacheable
		if req.cacheKey != nil && upstreamResp.Result != nil {
			if err = p.CacheFn(req, upstreamResp.Result); err != nil {
				req.logger.Error("failed to cache result", zap.Error(err))
			}
		}

		// remap request id
		resp := jsonrpc.JsonRpcResponse{
			ID:             req.JsonRpcRequest.ID,
			JsonRpcVersion: upstreamResp.JsonRpcVersion,
			Error:          upstreamResp.Error,
			Result:         upstreamResp.Result,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			req.logger.Error("failed to marshal resp", zap.Error(err))
			return
		}
		u.client.send <- RespData{Data: data}
	}
}
