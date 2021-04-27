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
	"github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
	restdidclient "github.com/trustbloc/agent-sdk/pkg/controller/rest/didclient"
)

func getDIDClient(t *testing.T) *DIDClient {
	t.Helper()

	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetDIDClient()
	require.NoError(t, err)
	require.NotNil(t, controller)

	dc, ok := controller.(*DIDClient)
	require.Equal(t, ok, true)

	return dc
}

func TestDIDClient_CreatePeerDID(t *testing.T) {
	dc := getDIDClient(t)

	response, err := json.Marshal(didclient.CreateDIDResponse{})
	require.NoError(t, err)

	dc.httpClient = &mockHTTPClient{
		data:   string(response),
		method: http.MethodPost, url: mockAgentURL + restdidclient.CreatePeerDIDPath,
	}

	payload, err := json.Marshal(didclient.CreatePeerDIDRequest{})
	require.NoError(t, err)

	resp := dc.CreatePeerDID(&models.RequestEnvelope{Payload: payload})

	require.NotNil(t, resp)
	require.Nil(t, resp.Error)
	require.Equal(t, string(response), string(resp.Payload))
}

func TestDIDClient_CreateTrustBlocDID(t *testing.T) {
	dc := getDIDClient(t)

	response, err := json.Marshal(didclient.CreateDIDResponse{})
	require.NoError(t, err)

	dc.httpClient = &mockHTTPClient{
		data:   string(response),
		method: http.MethodPost, url: mockAgentURL + restdidclient.CreateBlocDIDPath,
	}

	payload, err := json.Marshal(didclient.CreateBlocDIDRequest{})
	require.NoError(t, err)

	resp := dc.CreateTrustBlocDID(&models.RequestEnvelope{Payload: payload})

	require.NotNil(t, resp)
	require.Nil(t, resp.Error)
	require.Equal(t, string(response), string(resp.Payload))
}
