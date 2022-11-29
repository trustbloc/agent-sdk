/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest //nolint:testpackage // uses internal implementation details

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest/mediatorclient"
)

const (
	sampleConnectRequest = `{
    	"invitation": {
        	"@id": "3ae3d2cb-83bf-429f-93ea-0802f92ecf42",
        	"@type": "https://didcomm.org/out-of-band/1.0/invitation",
        	"label": "hub-router",
        	"service": [{
            	"ID": "1d03b636-ab0d-4a4e-904b-cdc70265c6bc",
            	"Type": "did-communication",
            	"Priority": 0,
            	"RecipientKeys": ["36umoSWgaY4pBpwGUX9UNXBmpo1iDSdLsiKDs4XPXK4Q"],
            	"RoutingKeys": null,
            	"ServiceEndpoint": "wss://hub.router.agent.example.com:10072",
            	"Properties": null
        	}],
        	"protocols": ["https://didcomm.org/didexchange/1.0"]
    	},
		"mylabel": "sample-agent-label"
		}`
	sampleCreateInvitationRequest = `{"label":"sample-abc"}`
	sampleSendConnectionRequest   = `{"didDoc": {"@id": "sample-did-id"}}`
)

func getMediatorClientController(t *testing.T) *MediatorClient {
	t.Helper()

	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetMediatorClientController()
	require.NoError(t, err)
	require.NotNil(t, controller)

	m, ok := controller.(*MediatorClient)
	require.Equal(t, ok, true)

	return m
}

func TestMediatorClient_Connect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		controller := getMediatorClientController(t)

		reqData := sampleConnectRequest
		mockResponse := `{"connectionID":"123-abc", "routerEndpoint":"test-ep", 
							"routingKeys": ["routingKey#1", "routingKey#2"]}`

		controller.httpClient = &mockHTTPClient{
			data:   mockResponse,
			method: http.MethodPost, url: mockAgentURL + mediatorclient.ConnectPath,
		}

		req := &models.RequestEnvelope{Payload: []byte(reqData)}
		resp := controller.Connect(req)

		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t, mockResponse, string(resp.Payload))
	})
}

func TestMediatorClient_CreateInvitation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		controller := getMediatorClientController(t)

		reqData := sampleCreateInvitationRequest
		mockResponse := `{"invitation":{"@id":"sample-id"}}`

		controller.httpClient = &mockHTTPClient{
			data:   mockResponse,
			method: http.MethodPost, url: mockAgentURL + mediatorclient.CreateInvitationPath,
		}

		req := &models.RequestEnvelope{Payload: []byte(reqData)}
		resp := controller.CreateInvitation(req)

		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t, mockResponse, string(resp.Payload))
	})
}

func TestMediatorClient_SendCreateConnectionRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		controller := getMediatorClientController(t)

		reqData := sampleSendConnectionRequest
		mockResponse := `{"payload":{"@id":"sample-id"}}`

		controller.httpClient = &mockHTTPClient{
			data:   mockResponse,
			method: http.MethodPost, url: mockAgentURL + mediatorclient.SendCreateConnectionRequest,
		}

		req := &models.RequestEnvelope{Payload: []byte(reqData)}
		resp := controller.SendCreateConnectionRequest(req)

		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t, mockResponse, string(resp.Payload))
	})
}
