//go:build js && wasm
// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main // nolint: testpackage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall/js"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// test callbacks.
var callbacks = make(map[string]chan *result) // nolint:gochecknoglobals

func TestMain(m *testing.M) {
	isTest = true

	go main()

	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		panic(errors.New("go main() timed out"))
	}

	results := make(chan *result)

	js.Global().Set("handleResult", js.FuncOf(acceptResults(results)))

	go dispatchResults(results)
	os.Exit(m.Run())
}

func TestEchoCmd(t *testing.T) {
	echo := newCommand("test", "echo", map[string]interface{}{"id": uuid.New().String()})
	result := make(chan *result)

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

func TestCreateVDRs(t *testing.T) {
	tests := []struct {
		name              string
		resolvers         []string
		blocDomain        string
		trustblocResolver string
		expected          int
		accept            map[int][]string
	}{{
		name: "Empty data",
		// expects default trustbloc resolver
		accept:   map[int][]string{0: {"orb"}},
		expected: 2,
	}, {
		name:      "Groups methods by resolver",
		resolvers: []string{"orb@http://resolver.com", "v1@http://resolver.com"},
		accept:    map[int][]string{0: {"orb", "v1"}, 1: {"orb"}},
		// expects resolver.com that supports trustbloc,v1 methods and default trustbloc resolver
		expected: 3,
	}, {
		name:      "Two different resolvers",
		resolvers: []string{"orb@http://resolver1.com", "v1@http://resolver2.com"},
		accept:    map[int][]string{0: {"orb"}, 1: {"v1"}, 2: {"orb"}},
		// expects resolver1.com and resolver2.com that supports trustbloc and v1 methods and default trustbloc resolver
		expected: 4,
	}}

	for _, test := range tests {
		res, err := createVDRs(test.resolvers, test.blocDomain, 10)

		for i, methods := range test.accept {
			for _, method := range methods {
				require.True(t, res[i].Accept(method))
			}
		}

		require.NoError(t, err)
		require.Equal(t, test.expected, len(res))
	}
}

func TestErrorCmd(t *testing.T) {
	errCmd := newCommand("test", "throwError", map[string]interface{}{})
	result := make(chan *result)
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

func TestCapabilityInvocationAction(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			path           string
			method         string
			expectedAction string
		}{
			{"/v1/keystores/{key_store_id}/keys", http.MethodPost, "createKey"},
			{"/v1/keystores/{key_store_id}/keys", http.MethodPut, "importKey"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}", http.MethodGet, "exportKey"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/sign", http.MethodPost, "sign"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/verify", http.MethodPost, "verify"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/encrypt", http.MethodPost, "encrypt"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/decrypt", http.MethodPost, "decrypt"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/computemac", http.MethodPost, "computeMAC"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/verifymac", http.MethodPost, "verifyMAC"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/signmulti", http.MethodPost, "signMulti"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/verifymulti", http.MethodPost, "verifyMulti"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/deriveproof", http.MethodPost, "deriveProof"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/verifyproof", http.MethodPost, "verifyProof"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/easy", http.MethodPost, "easy"},
			{"/v1/keystores/{key_store_id}/easyopen", http.MethodPost, "easyOpen"},
			{"/v1/keystores/{key_store_id}/sealopen", http.MethodPost, "sealOpen"},
			{"/v1/keystores/{key_store_id}/wrap", http.MethodPost, "wrap"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/wrap", http.MethodPost, "wrap"},
			{"/v1/keystores/{key_store_id}/keys/{key_id}/unwrap", http.MethodPost, "unwrap"},
		}

		for _, tt := range tests {
			t.Run(tt.expectedAction, func(t *testing.T) {
				t.Parallel()

				req := &http.Request{
					URL:    &url.URL{Path: tt.path},
					Method: tt.method,
				}

				action, err := capabilityInvocationAction(req)

				require.NoError(t, err)
				require.Equal(t, tt.expectedAction, action)
			})
		}
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			path        string
			method      string
			expectedErr string
		}{
			{"/v1/keystores/{key_store_id}/keys/{key_id}/sign", http.MethodGet, "unsupported operation: GET /sign"},
			{"/v1/keystores/{key_store_id}/keys", http.MethodGet, "unsupported operation: GET /keys"},
			{"/v1/keystores/{key_store_id}", http.MethodPost, "invalid path"},
			{"/v1/keystores/did", http.MethodPost, "invalid path"},
			{"/v1/keystores", http.MethodPost, "invalid path"},
			{"", http.MethodGet, "invalid path"},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s%s", tt.method, strings.ReplaceAll(tt.path, "/", "_")), func(t *testing.T) {
				t.Parallel()

				req := &http.Request{
					URL:    &url.URL{Path: tt.path},
					Method: tt.method,
				}

				action, err := capabilityInvocationAction(req)

				require.Empty(t, action)
				require.EqualError(t, err, tt.expectedErr)
			})
		}
	})
}

func acceptResults(in chan *result) func(js.Value, []js.Value) interface{} {
	return func(_ js.Value, args []js.Value) interface{} {
		r := &result{}
		if err := json.Unmarshal([]byte(args[0].String()), r); err != nil {
			panic(err)
		}
		in <- r

		return nil
	}
}

func dispatchResults(in chan *result) {
	for r := range in {
		cb, found := callbacks[r.ID]
		if !found {
			panic(fmt.Errorf("callback with ID %s not found", r.ID))
		}
		cb <- r
	}
}

func newCommand(pkg, fn string, payload map[string]interface{}) *command {
	return &command{
		ID:      uuid.New().String(),
		Pkg:     pkg,
		Fn:      fn,
		Payload: payload,
	}
}

func toString(c *command) string {
	s, err := json.Marshal(c)
	if err != nil {
		panic(fmt.Errorf("failed to marshal: %+v", c))
	}

	return string(s)
}
