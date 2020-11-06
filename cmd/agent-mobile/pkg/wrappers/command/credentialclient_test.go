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
	"github.com/trustbloc/agent-sdk/pkg/controller/command/credentialclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
)

func getCredentialClient(t *testing.T) *CredentialClient {
	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetCredentialClient()
	require.NoError(t, err)
	require.NotNil(t, controller)

	client, ok := controller.(*CredentialClient)
	require.Equal(t, ok, true)

	return client
}

func TestCredentialClient_SaveCredential(t *testing.T) {
	t.Run("save credential", func(t *testing.T) {
		client := getCredentialClient(t)

		fakeHandler := mockCommandRunner{data: nil}
		client.handlers[credentialclient.SaveCredentialCommandMethod] = fakeHandler.exec
		require.Len(t, client.handlers, 1)

		payload, err := json.Marshal(sdscomm.SaveCredentialToSDSRequest{})
		require.NoError(t, err)

		req := &models.RequestEnvelope{Payload: payload}
		resp := client.SaveCredential(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)

		require.Equal(t, []byte(nil), resp.Payload)
	})

	t.Run("custom error", func(t *testing.T) {
		client := getCredentialClient(t)

		client.handlers[credentialclient.SaveCredentialCommandMethod] = func(rw io.Writer, req io.Reader) command.Error {
			return command.NewExecuteError(1, errors.New("error"))
		}

		payload, err := json.Marshal(sdscomm.SaveCredentialToSDSRequest{})
		require.NoError(t, err)

		req := &models.RequestEnvelope{Payload: payload}
		resp := client.SaveCredential(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)

		require.Equal(t, &models.CommandError{Message: "error", Code: 1, Type: 1}, resp.Error)
	})

	t.Run("JSON error", func(t *testing.T) {
		client := getCredentialClient(t)

		req := &models.RequestEnvelope{Payload: []byte(`{`)}
		resp := client.SaveCredential(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		require.Equal(t, "unexpected end of JSON input", resp.Error.Message)
	})
}
