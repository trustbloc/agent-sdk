/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest // nolint:testpackage // uses internal implementation details

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
	restcredentialclient "github.com/trustbloc/agent-sdk/pkg/controller/rest/credentialclient"
)

func getCredentialClient(t *testing.T) *CredentialClient {
	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetCredentialClient()
	require.NoError(t, err)
	require.NotNil(t, controller)

	cc, ok := controller.(*CredentialClient)
	require.Equal(t, ok, true)

	return cc
}

func TestCredentialClient_SaveCredential(t *testing.T) {
	cc := getCredentialClient(t)

	cc.httpClient = &mockHTTPClient{
		method: http.MethodPost, url: mockAgentURL + restcredentialclient.SaveCredentialPath,
	}

	payload, err := json.Marshal(sdscomm.SaveCredentialToSDSRequest{})
	require.NoError(t, err)

	resp := cc.SaveCredential(&models.RequestEnvelope{Payload: payload})

	require.NotNil(t, resp)
	require.Nil(t, resp.Error)
	require.Equal(t, []byte{}, resp.Payload)
}
