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
	"github.com/trustbloc/agent-sdk/pkg/controller/command/presentationclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
)

func getPresentationClient(t *testing.T) *PresentationClient {
	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetPresentationClient()
	require.NoError(t, err)
	require.NotNil(t, controller)

	client, ok := controller.(*PresentationClient)
	require.Equal(t, ok, true)

	return client
}

func TestPresentationClient_SavePresentation(t *testing.T) {
	t.Run("save presentation", func(t *testing.T) {
		client := getPresentationClient(t)

		fakeHandler := mockCommandRunner{data: nil}
		client.handlers[presentationclient.SavePresentationCommandMethod] = fakeHandler.exec
		require.Len(t, client.handlers, 1)

		payload, err := json.Marshal(sdscomm.SavePresentationToSDSRequest{})
		require.NoError(t, err)

		req := &models.RequestEnvelope{Payload: payload}
		resp := client.SavePresentation(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)

		require.Equal(t, []byte(nil), resp.Payload)
	})

	t.Run("custom error", func(t *testing.T) {
		client := getPresentationClient(t)

		client.handlers[presentationclient.SavePresentationCommandMethod] = func(rw io.Writer, req io.Reader) command.Error {
			return command.NewExecuteError(1, errors.New("error"))
		}

		payload, err := json.Marshal(sdscomm.SavePresentationToSDSRequest{})
		require.NoError(t, err)

		req := &models.RequestEnvelope{Payload: payload}
		resp := client.SavePresentation(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)

		require.Equal(t, &models.CommandError{Message: "error", Code: 1, Type: 1}, resp.Error)
	})

	t.Run("JSON error", func(t *testing.T) {
		client := getPresentationClient(t)

		req := &models.RequestEnvelope{Payload: []byte(`{`)}
		resp := client.SavePresentation(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		require.Equal(t, "unexpected end of JSON input", resp.Error.Message)
	})
}
