//go:build js && wasm
// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wasmsetup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"syscall/js"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/google/uuid"
	controllercmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/trustbloc/edge-core/pkg/log"
)

var logger = log.New("agent-js-worker")

const (
	wasmStartupTopic = "asset-ready"
	handleResultFn   = "handleResult"
	commandPkg       = "agent"
	startFn          = "Start"
	stopFn           = "Stop"
)

// TODO Signal JS when WASM is loaded and ready.
//      This is being used in tests for now.
var (
	ready  = make(chan struct{}) //nolint:gochecknoglobals
	isTest = false               //nolint:gochecknoglobals
)

// Command is received from JS.
type Command struct {
	ID      string                 `json:"id"`
	Pkg     string                 `json:"pkg"`
	Fn      string                 `json:"fn"`
	Payload map[string]interface{} `json:"payload"`
}

// Result is sent back to JS.
type Result struct {
	ID      string                 `json:"id"`
	IsErr   bool                   `json:"isErr"`
	ErrMsg  string                 `json:"errMsg"`
	Payload map[string]interface{} `json:"payload"`
	Topic   string                 `json:"topic"`
}

type AddAgentHandlers func(pkgMap map[string]map[string]func(*Command) *Result)

func TakeFrom(in chan *Command) func(js.Value, []js.Value) interface{} {
	return func(_ js.Value, args []js.Value) interface{} {
		cmd := &Command{}
		if err := json.Unmarshal([]byte(args[0].String()), cmd); err != nil {
			logger.Errorf("agent wasm: unable to unmarshal input=%s. err=%s", args[0].String(), err)

			return nil
		}

		in <- cmd

		return nil
	}
}

func Pipe(input chan *Command, output chan *Result, addAgentHandlers AddAgentHandlers, workers int) {
	handlers := testHandlers()

	addAgentHandlers(handlers)

	// Upon the first call `btcec.S256()` deserializes the pre-computed byte points for the secp256k1 curve and
	// it takes some time. Triggering that function here speeds up the following protocols.
	go initS256()

	for w := 0; w < workers; w++ {
		go worker(input, output, handlers)
	}
}

func initS256() {
	btcec.S256()
}

func worker(input chan *Command, output chan *Result, handlers map[string]map[string]func(*Command) *Result) {
	for c := range input {
		if c.ID == "" {
			logger.Warnf("agent wasm: missing ID for input: %v", c)
		}

		if pkg, found := handlers[c.Pkg]; found {
			if fn, found := pkg[c.Fn]; found {
				output <- fn(c)

				continue
			}
		}

		output <- handlerNotFoundErr(c)
	}
}

func SendTo(out chan *Result) {
	for r := range out {
		out, err := json.Marshal(r)
		if err != nil {
			logger.Errorf("agent wasm: failed to marshal response for id=%s err=%s ", r.ID, err)
		}

		js.Global().Call(handleResultFn, string(out))
	}
}

func testHandlers() map[string]map[string]func(*Command) *Result {
	return map[string]map[string]func(*Command) *Result{
		"test": {
			"echo": func(c *Command) *Result {
				return &Result{
					ID:      c.ID,
					Payload: map[string]interface{}{"echo": c.Payload},
				}
			},
			"throwError": func(c *Command) *Result {
				return NewErrResult(c.ID, "an error !!")
			},
			"timeout": func(c *Command) *Result {
				const echoTimeout = 10 * time.Second

				time.Sleep(echoTimeout)

				return &Result{
					ID:      c.ID,
					Payload: map[string]interface{}{"echo": c.Payload},
				}
			},
		},
	}
}

func isStartCommand(c *Command) bool {
	return c.Pkg == commandPkg && c.Fn == startFn
}

func isStopCommand(c *Command) bool {
	return c.Pkg == commandPkg && c.Fn == stopFn
}

func handlerNotFoundErr(c *Command) *Result {
	if isStartCommand(c) {
		return NewErrResult(c.ID, "Agent already started")
	} else if isStopCommand(c) {
		return NewErrResult(c.ID, "Agent not running")
	}

	return NewErrResult(c.ID, fmt.Sprintf("invalid pkg/fn: %s/%s, make sure agent is started", c.Pkg, c.Fn))
}

// JSNotifier notifies about all incoming events.
type JSNotifier struct{}

// Notify is mock implementation of webhook notifier Notify().
func (n *JSNotifier) Notify(topic string, message []byte) error {
	payload := make(map[string]interface{})
	if err := json.Unmarshal(message, &payload); err != nil {
		return err
	}

	out, err := json.Marshal(&Result{
		ID:      uuid.New().String(),
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		return err
	}

	js.Global().Call(handleResultFn, string(out))

	return nil
}

func PostInitMsg() {
	if isTest {
		return
	}

	out, err := json.Marshal(&Result{
		ID:    uuid.New().String(),
		Topic: wasmStartupTopic,
	})
	if err != nil {
		panic(err)
	}

	js.Global().Call(handleResultFn, string(out))
}

type execFn func(rw io.Writer, req io.Reader) error

type commandHandler struct {
	name   string
	method string
	exec   execFn
}

func wrapAriesHandler(handlers []controllercmd.Handler) []commandHandler {
	var hh []commandHandler

	for _, h := range handlers {
		handle := h.Handle()

		hh = append(hh, commandHandler{
			name:   h.Name(),
			method: h.Method(),
			exec: func(rw io.Writer, req io.Reader) error {
				e := handle(rw, req)
				if e != nil {
					return fmt.Errorf("code: %+v, message: %w", e.Code(), e)
				}

				return nil
			},
		})
	}

	return hh
}

func AddCommandHandlers(handlers []controllercmd.Handler, pkgMap map[string]map[string]func(*Command) *Result) {
	for _, h := range wrapAriesHandler(handlers) {
		fnMap, ok := pkgMap[h.name]
		if !ok {
			fnMap = make(map[string]func(*Command) *Result)
		}

		fnMap[h.method] = cmdExecToFn(h.exec)
		pkgMap[h.name] = fnMap
	}
}

func cmdExecToFn(exec execFn) func(*Command) *Result {
	return func(c *Command) *Result {
		b, er := json.Marshal(c.Payload)
		if er != nil {
			return &Result{
				ID:     c.ID,
				IsErr:  true,
				ErrMsg: fmt.Sprintf("agent wasm: failed to unmarshal payload. err=%s", er),
			}
		}

		req := bytes.NewBuffer(b)

		var buf bytes.Buffer

		err := exec(&buf, req)
		if err != nil {
			return NewErrResult(c.ID, err.Error())
		}

		payload := make(map[string]interface{})

		if len(buf.Bytes()) > 0 {
			if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
				return &Result{
					ID:     c.ID,
					IsErr:  true,
					ErrMsg: fmt.Sprintf("agent wasm: failed to unmarshal Command Result=%+v err=%s", buf.String(), err),
				}
			}
		}

		return &Result{
			ID:      c.ID,
			Payload: payload,
		}
	}
}

func NewErrResult(id, msg string) *Result {
	return &Result{
		ID:     id,
		IsErr:  true,
		ErrMsg: "agent wasm: " + msg,
	}
}
