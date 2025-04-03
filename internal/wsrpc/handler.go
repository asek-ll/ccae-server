package wsrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/asek-ll/aecc-server/internal/ws"
	"github.com/asek-ll/aecc-server/pkg/gopool"
)

var _ ws.Handler = &JsonRpcServer{}

func Typed[T any](f func(clientId uint, params T) (any, error)) RpcMethod {
	return func(clientId uint, params []byte) (any, error) {
		var ps T
		err := json.Unmarshal(params, &ps)
		if err != nil {
			return nil, err
		}

		return f(clientId, ps)
	}
}

type RpcMethod = func(clientId uint, params []byte) (any, error)

type ClientCtx struct {
	WsConnectionId uint
	InnerId        string
}

type JsonRpcServer struct {
	methods           map[string]RpcMethod
	wsServer          *ws.Server
	reqSeqMu          sync.RWMutex
	reqSeq            uint
	pending           map[uint]chan *Message
	pool              *gopool.Pool
	disconnectHandler func(clientId uint) error
}

func NewServer(server *ws.Server) *JsonRpcServer {
	rpcServer := &JsonRpcServer{
		methods:  make(map[string]RpcMethod),
		wsServer: server,
		pending:  make(map[uint]chan *Message),
		pool:     gopool.NewPool(128, 1, 1),
	}

	server.SetHandler(rpcServer)

	return rpcServer
}

func (h *JsonRpcServer) SetDisconnectHandler(handler func(clientId uint) error) {
	h.disconnectHandler = handler
}

func (h *JsonRpcServer) AddMethod(name string, method RpcMethod) {
	h.methods[name] = method
}

func (h *JsonRpcServer) HandleMessage(content []byte, client *ws.Client) error {
	var msg Message
	err := json.Unmarshal(content, &msg)
	log.Println("[DEBUG] Received: ", msg)
	if err != nil {
		return nil
	}

	if msg.Method != nil && msg.Result == nil && msg.Error == nil {
		m, e := h.methods[*msg.Method]
		if !e {
			log.Printf("[WARN] Unknown method: %s", *msg.Method)
			return nil
		}

		h.pool.Schedule(func() {
			res, err := m(client.ID(), msg.Params)
			var response *Response
			if err != nil {
				response = &Response{
					JsonRpc: "2.0",
					ID:      msg.ID,
					Error: &Error{
						Code:    -1,
						Message: err.Error(),
					},
				}
			} else if res != nil {
				response = &Response{
					JsonRpc: "2.0",
					ID:      msg.ID,
					Result:  res,
				}
			}
			if response != nil {
				err := client.WriteJSON(*response)
				if err != nil {
					log.Printf("[ERROR] Error on method handle: %s, %v", *msg.Method, err)
				}
			}
		})
		return nil
	}

	if msg.Method == nil && (msg.Result != nil || msg.Error != nil) {
		h.reqSeqMu.RLock()
		done, e := h.pending[msg.ID]
		h.reqSeqMu.RUnlock()
		if !e {
			log.Printf("[ERROR] Not found pendition for request id: %d", msg.ID)
			return nil
		}
		h.pool.Schedule(func() {
			done <- &msg
		})

		return nil
	}

	return nil
}

func (h *JsonRpcServer) SendRequest(clientId uint, method string, params any) (uint, error) {
	client, e := h.wsServer.GetClient(clientId)
	if !e {
		return 0, errors.New("Client not exists")
	}
	h.reqSeqMu.Lock()
	reqId := h.reqSeq
	h.reqSeq = h.reqSeq % 1000000
	h.reqSeqMu.Unlock()

	request := Request{
		JsonRpc: "2.0",
		ID:      reqId,
		Method:  method,
		Params:  params,
	}

	err := client.WriteJSON(request)

	if err != nil {
		return 0, err
	}

	return reqId, nil
}

func (h *JsonRpcServer) SendRequestSync(ctx context.Context, clientId uint, method string, params any, result interface{}) error {
	client, e := h.wsServer.GetClient(clientId)
	if !e {
		return errors.New("Client not exists")
	}
	log.Printf("[INFO] Client call %d %s %v", clientId, method, params)

	done := make(chan *Message)

	h.reqSeqMu.Lock()
	reqId := h.reqSeq
	h.reqSeq++
	h.pending[reqId] = done
	h.reqSeqMu.Unlock()

	request := Request{
		JsonRpc: "2.0",
		ID:      reqId,
		Method:  method,
		Params:  params,
	}

	err := client.WriteJSON(request)
	if err != nil {
		h.reqSeqMu.Lock()
		delete(h.pending, reqId)
		h.reqSeqMu.Unlock()
		return err
	}

	select {
	case <-ctx.Done():
		h.reqSeqMu.Lock()
		delete(h.pending, reqId)
		h.reqSeqMu.Unlock()
		return errors.New("Timeout")
	case msg := <-done:

		if msg.Result != nil {
			err := json.Unmarshal(msg.Result, result)
			if err != nil {
				return err
			}
			return nil
		}

		if msg.Error != nil {
			return fmt.Errorf("Got error: %v", msg.Error)
		}
	}

	return nil
}

func (h *JsonRpcServer) HandleDisconnect(client *ws.Client) {
	if h.disconnectHandler != nil {
		err := h.disconnectHandler(client.ID())
		if err != nil {
			log.Printf("Error on disconnect %d", client.ID())
		}
	}
}
