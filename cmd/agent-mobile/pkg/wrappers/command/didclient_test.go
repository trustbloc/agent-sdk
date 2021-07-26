/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command // nolint:testpackage // uses internal implementation details

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
)

func getDIDClient(t *testing.T) *DIDClient {
	t.Helper()

	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetDIDClient()
	require.NoError(t, err)
	require.NotNil(t, controller)

	client, ok := controller.(*DIDClient)
	require.Equal(t, ok, true)

	return client
}

func TestDIDClient_CreateOrbDID(t *testing.T) {
	t.Run("creates trust bloc DID", func(t *testing.T) {
		client := getDIDClient(t)

		response, err := json.Marshal(didclient.CreateDIDResponse{})
		require.NoError(t, err)

		fakeHandler := mockCommandRunner{data: response}
		client.handlers[didclient.CreateOrbDIDCommandMethod] = fakeHandler.exec

		payload, err := json.Marshal(didclient.CreateOrbDIDRequest{})
		require.NoError(t, err)

		req := &models.RequestEnvelope{Payload: payload}
		resp := client.CreateOrbDID(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)

		require.Equal(t, string(response), string(resp.Payload))
	})

	t.Run("custom error", func(t *testing.T) {
		client := getDIDClient(t)

		client.handlers[didclient.CreateOrbDIDCommandMethod] = func(rw io.Writer, req io.Reader) command.Error {
			return command.NewExecuteError(1, errors.New("error"))
		}

		payload, err := json.Marshal(didclient.CreateOrbDIDRequest{})
		require.NoError(t, err)

		req := &models.RequestEnvelope{Payload: payload}
		resp := client.CreateOrbDID(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)

		require.Equal(t, &models.CommandError{Message: "error", Code: 1, Type: 1}, resp.Error)
	})

	t.Run("JSON error", func(t *testing.T) {
		client := getDIDClient(t)

		req := &models.RequestEnvelope{Payload: []byte(`{`)}
		resp := client.CreateOrbDID(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		require.Equal(t, "unexpected end of JSON input", resp.Error.Message)
	})
}

func TestDIDClient_CreatePeerDID(t *testing.T) {
	t.Run("creates peer DID", func(t *testing.T) {
		client := getDIDClient(t)

		response, err := json.Marshal(didclient.CreateDIDResponse{})
		require.NoError(t, err)

		fakeHandler := mockCommandRunner{data: response}
		client.handlers[didclient.CreatePeerDIDCommandMethod] = fakeHandler.exec

		payload, err := json.Marshal(didclient.CreatePeerDIDRequest{})
		require.NoError(t, err)

		req := &models.RequestEnvelope{Payload: payload}
		resp := client.CreatePeerDID(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)

		require.Equal(t, string(response), string(resp.Payload))
	})

	t.Run("custom error", func(t *testing.T) {
		client := getDIDClient(t)

		client.handlers[didclient.CreatePeerDIDCommandMethod] = func(rw io.Writer, req io.Reader) command.Error {
			return command.NewExecuteError(1, errors.New("error"))
		}

		payload, err := json.Marshal(didclient.CreateOrbDIDRequest{})
		require.NoError(t, err)

		req := &models.RequestEnvelope{Payload: payload}
		resp := client.CreatePeerDID(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)

		require.Equal(t, &models.CommandError{Message: "error", Code: 1, Type: 1}, resp.Error)
	})

	t.Run("JSON error", func(t *testing.T) {
		client := getDIDClient(t)

		req := &models.RequestEnvelope{Payload: []byte(`{`)}
		resp := client.CreatePeerDID(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		require.Equal(t, "unexpected end of JSON input", resp.Error.Message)
	})
}
