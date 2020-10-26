/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package cmdutil_test

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/internal/cmdutil"
)

func TestNewHTTPHandler(t *testing.T) {
	path := "/sample-path"
	method := "GET"
	handled := make(chan bool)
	handlerFn := func(w http.ResponseWriter, r *http.Request) {
		// do nothing
		handled <- true
	}

	handler := cmdutil.NewHTTPHandler(path, method, handlerFn)
	require.Equal(t, path, handler.Path())
	require.Equal(t, method, handler.Method())
	require.NotNil(t, handler.Handle())

	go handler.Handle()(nil, nil)

	select {
	case res := <-handled:
		require.True(t, res)
	case <-time.After(2 * time.Second):
		t.Fatal("handler function didnt get executed")
	}
}

func TestNewCommandHandler(t *testing.T) {
	name := "foo"
	method := "bar"
	handled := make(chan bool)
	handlerFn := func(rw io.Writer, req io.Reader) command.Error {
		// do nothing
		handled <- true

		return nil
	}

	handler := cmdutil.NewCommandHandler(name, method, handlerFn)
	require.Equal(t, name, handler.Name())
	require.Equal(t, method, handler.Method())
	require.NotNil(t, handler.Handle())

	go func() {
		err := handler.Handle()(nil, nil)
		require.NoError(t, err)
	}()

	select {
	case res := <-handled:
		require.True(t, res)
	case <-time.After(2 * time.Second):
		t.Fatal("handler function didnt get executed")
	}
}
