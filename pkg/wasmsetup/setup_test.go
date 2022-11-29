//go:build js && wasm
// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wasmsetup_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/trustbloc/agent-sdk/pkg/wasmsetup"
)

// test callbacks.
var callbacks = make(map[string]chan *wasmsetup.Result) //nolint:gochecknoglobals

func TestSetup(t *testing.T) {
	results := make(chan *wasmsetup.Result)
	js.Global().Set("handleResult", js.FuncOf(acceptResults(results)))

	go setup()

	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		panic(errors.New("go main() timed out"))
	}

	go dispatchResults(results)
}

func TestEchoCmd(t *testing.T) {
	echo := newCommand("test", "echo", map[string]interface{}{"id": uuid.New().String()})
	result := make(chan *wasmsetup.Result)

	callbacks[echo.ID] = result
	defer delete(callbacks, echo.ID)

	js.Global().Call("handleMsg", toString(echo))

	select {
	case r := <-result:
		assert.Equal(t, echo.ID, r.ID)
		assert.False(t, r.IsErr)
		assert.Empty(t, r.ErrMsg)
		assert.Equal(t, r.Payload["echo"], echo.Payload)
	case <-time.After(5 * time.Second):
		t.Error("test timeout")
	}
}

func TestErrorCmd(t *testing.T) {
	errCmd := newCommand("test", "throwError", map[string]interface{}{})
	result := make(chan *wasmsetup.Result)
	callbacks[errCmd.ID] = result

	defer delete(callbacks, errCmd.ID)

	js.Global().Call("handleMsg", toString(errCmd))

	select {
	case r := <-result:
		assert.Equal(t, errCmd.ID, r.ID)
		assert.True(t, r.IsErr)
		assert.NotEmpty(t, r.ErrMsg)
		assert.Empty(t, r.Payload)
	case <-time.After(5 * time.Second):
		t.Error("test timeout")
	}
}

func newCommand(pkg, fn string, payload map[string]interface{}) *wasmsetup.Command {
	return &wasmsetup.Command{
		ID:      uuid.New().String(),
		Pkg:     pkg,
		Fn:      fn,
		Payload: payload,
	}
}

func toString(c *wasmsetup.Command) string {
	s, err := json.Marshal(c)
	if err != nil {
		panic(fmt.Errorf("failed to marshal: %+v", c))
	}

	return string(s)
}

var (
	ready  = make(chan struct{}) //nolint:gochecknoglobals
	isTest = false               //nolint:gochecknoglobals
)

func setup() {
	input := make(chan *wasmsetup.Command, 10) //nolint: gomnd
	output := make(chan *wasmsetup.Result)

	go wasmsetup.Pipe(input, output, func(pkgMap map[string]map[string]func(*wasmsetup.Command) *wasmsetup.Result) {

	}, 2)

	go wasmsetup.SendTo(output)

	js.Global().Set("handleMsg", js.FuncOf(wasmsetup.TakeFrom(input)))

	wasmsetup.PostInitMsg()

	if isTest {
		ready <- struct{}{}
	}

	select {}
}

func acceptResults(in chan *wasmsetup.Result) func(js.Value, []js.Value) interface{} {
	return func(_ js.Value, args []js.Value) interface{} {
		r := &wasmsetup.Result{}
		if err := json.Unmarshal([]byte(args[0].String()), r); err != nil {
			panic(err)
		}
		in <- r

		return nil
	}
}

func dispatchResults(in chan *wasmsetup.Result) {
	for r := range in {
		cb, found := callbacks[r.ID]
		if !found {
			panic(fmt.Errorf("callback with ID %s not found", r.ID))
		}
		cb <- r
	}
}
