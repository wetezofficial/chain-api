/*
 * Created by Du, Chengbin on 2022/4/27.
 */

package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"starnet/chain-api/pkg/jsonrpc"
	"sync"
	"time"
)

type Client struct {
	conn   *websocket.Conn
	send   chan<- RespData
	closed bool
	mutex  *sync.Mutex
}

func NewClient(conn *websocket.Conn, send chan<- RespData) *Client {
	return &Client{conn: conn, send: send, mutex: new(sync.Mutex)}
}

func (c *Client) SetClosed() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.closed = true
}

func (c *Client) Send(data RespData) {
	c.mutex.Lock()
	if c.closed {
		c.mutex.Unlock()
		return
	}
	c.mutex.Unlock()

	c.send <- data
}

type UpstreamWebSocket struct {
	conn   *websocket.Conn
	client *Client
	proxy  *JsonRpcProxy
	logger *zap.Logger

	mutex    *sync.Mutex
	requests map[uint64]*request
}

func (u *UpstreamWebSocket) Close() error {
	u.client.SetClosed()
	return u.conn.Close()
}

func (u *UpstreamWebSocket) Send(ctx context.Context, logger *zap.Logger, rawreq *jsonrpc.JsonRpcRequest) error {
	if rawreq.IsBatchCall() {
		return u.conn.WriteJSON(rawreq)
	}

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
		u.client.Send(RespData{Data: resp})
		return nil
	}

	u.mutex.Lock()
	u.requests[req.ID] = req
	u.mutex.Unlock()
	return u.conn.WriteJSON(req)
}

func (u *UpstreamWebSocket) run() {
	defer u.conn.Close()
	defer u.client.conn.Close()

	u.conn.SetPongHandler(func(appData string) error {
		return u.client.conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(10*time.Second))
	})

	u.client.conn.SetPingHandler(func(appData string) error {
		return u.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second))
	})

	p := u.proxy
	ws := u.conn
	for {
		_, rawresp, err := ws.ReadMessage()
		if err != nil {
			return
		}
		u.logger.Debug("got resp from upstream", zap.ByteString("rawresp", rawresp))

		rawresp = bytes.TrimSpace(rawresp)
		if rawresp[0] == '[' && rawresp[len(rawresp)-1] == ']' {
			// batch call response
			u.client.Send(RespData{Data: rawresp})
			continue
		}

		//req.logger.Debug("new upstream response", zap.ByteString("resp", rawreq))
		upstreamResp := UpstreamJsonRpcResponse{}
		if err = json.Unmarshal(rawresp, &upstreamResp); err != nil {
			return
		}

		// 订阅的通知是没有 id 字段的
		if upstreamResp.ID == 0 {
			// 直接把内容写入到客户端
			u.client.Send(RespData{Data: rawresp, Subscription: true})
			continue
		}

		u.mutex.Lock()
		req, ok := u.requests[upstreamResp.ID]
		u.mutex.Unlock()
		if ok {

			u.mutex.Lock()
			delete(u.requests, upstreamResp.ID)
			u.mutex.Unlock()

			// step3. Cache if it is a valid result and cacheable
			if req.cacheKey != nil && upstreamResp.Result != nil {
				if err = p.CacheFn(req, upstreamResp.Result); err != nil {
					req.logger.Error("failed to cache result", zap.Error(err))
				}
			}
		}

		u.client.Send(RespData{Data: rawresp})
	}
}
