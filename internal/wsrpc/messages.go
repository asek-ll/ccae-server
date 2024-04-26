package wsrpc

import "encoding/json"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Message struct {
	JsonRpc string          `json:"jsonrpc"`
	ID      uint            `json:"id"`
	Method  *string         `json:"method"`
	Params  json.RawMessage `json:"params"`
	Result  json.RawMessage `json:"result"`
	Error   *Error          `json:"error"`
}

type Response struct {
	JsonRpc string `json:"jsonrpc"`
	ID      uint   `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Request struct {
	JsonRpc string `json:"jsonrpc"`
	ID      uint   `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}
