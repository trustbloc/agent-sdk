/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package presentationclient_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest"
	. "github.com/trustbloc/agent-sdk/pkg/controller/rest/presentationclient"
)

func TestOperation_GetRESTHandlers(t *testing.T) {
	operation := New(sdscomm.New(""))
	require.Len(t, operation.GetRESTHandlers(), 1)
}

func TestOperation_AcceptProposal(t *testing.T) {
	t.Run("Bad request", func(t *testing.T) {
		operation := New(sdscomm.New(""))

		_, code, err := sendRequestToHandler(
			handlerLookup(t, operation, SavePresentationPath),
			bytes.NewBufferString(`{}`), SavePresentationPath)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, code)
	})
}

func handlerLookup(t *testing.T, op *Operation, lookup string) rest.Handler {
	t.Helper()

	handlers := op.GetRESTHandlers()
	require.NotEmpty(t, handlers)

	for _, h := range handlers {
		if h.Path() == lookup {
			return h
		}
	}

	require.Fail(t, "unable to find handler")

	return nil
}

// sendRequestToHandler reads response from given http handle func.
func sendRequestToHandler(handler rest.Handler, requestBody io.Reader, path string) (*bytes.Buffer, int, error) {
	// prepare request
	req, err := http.NewRequestWithContext(context.Background(), handler.Method(), path, requestBody)
	if err != nil {
		return nil, 0, err
	}

	// prepare router
	router := mux.NewRouter()

	router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())

	// create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// serve http on given response and request
	router.ServeHTTP(rr, req)

	return rr.Body, rr.Code, nil
}
