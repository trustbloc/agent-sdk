/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command // nolint:testpackage // uses internal implementation details

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/mediatorclient"
)

const (
	sampleConnectRequest = `{
    	"invitation": {
        	"@id": "3ae3d2cb-83bf-429f-93ea-0802f92ecf42",
        	"@type": "https://didcomm.org/oob-invitation/1.0/invitation",
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
		mediatorClientController := getMediatorClientController(t)

		mockResponse := `{"connectionID":"123-abc", "routerEndpoint":"test-ep", 
							"routingKeys": ["routingKey#1", "routingKey#2"]}`
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}

		mediatorClientController.handlers[mediatorclient.Connect] = fakeHandler.exec

		payload := sampleConnectRequest

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := mediatorClientController.Connect(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestMediatorClient_CreateInvitation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mediatorClientController := getMediatorClientController(t)

		mockResponse := `{"invitation":{"@id":"sample-id"}}`
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}

		mediatorClientController.handlers[mediatorclient.CreateInvitation] = fakeHandler.exec

		payload := sampleCreateInvitationRequest

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := mediatorClientController.CreateInvitation(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestMediatorClient_SendCreateConnectionRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mediatorClientController := getMediatorClientController(t)

		mockResponse := `{"payload":{"@id":"sample-id"}}`
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}

		mediatorClientController.handlers[mediatorclient.SendCreateConnectionRequest] = fakeHandler.exec

		payload := sampleSendConnectionRequest

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := mediatorClientController.SendCreateConnectionRequest(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}
