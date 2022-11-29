//go:build js && wasm
// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"errors"
	"syscall/js"
	"testing"
	"time"

	"github.com/trustbloc/agent-sdk/pkg/wasmsetup"
)

func TestMain(_ *testing.M) {
	isTest = true

	results := make(chan *wasmsetup.Result)

	js.Global().Set("handleResult", js.FuncOf(acceptResults(results)))

	go main()

	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		panic(errors.New("go main() timed out"))
	}
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
