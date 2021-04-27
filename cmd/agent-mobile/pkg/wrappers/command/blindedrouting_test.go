/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command // nolint:testpackage // uses internal implementation details

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/blindedrouting"
)

const (
	sampleDIDDocRequest        = `{"connectionID":"sample-conn-01"}`
	sampleRegisterRouteRequest = `{"messageID":"sample-msg-01", "didDoc":{"@id":"sample-did-id"}}`
)

func getBlindedRoutingController(t *testing.T) *BlindedRouting {
	t.Helper()

	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetBlindedRoutingController()
	require.NoError(t, err)
	require.NotNil(t, controller)

	m, ok := controller.(*BlindedRouting)
	require.Equal(t, ok, true)

	return m
}

func TestBlindedRouting_SendDIDDocRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		blindedRoutingController := getBlindedRoutingController(t)

		mockResponse := `{"payload": {"didDoc": {"@id":"sample-did-id"}}}`
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}

		blindedRoutingController.handlers[blindedrouting.SendDIDDocRequest] = fakeHandler.exec

		payload := sampleDIDDocRequest

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := blindedRoutingController.SendDIDDocRequest(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestBlindedRouting_SendRegisterRouteRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		blindedRoutingController := getBlindedRoutingController(t)

		mockResponse := `{"payload": {"message": {"@id":"sample-did-id"}}}`
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}

		blindedRoutingController.handlers[blindedrouting.SendRegisterRouteRequest] = fakeHandler.exec

		payload := sampleRegisterRouteRequest

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := blindedRoutingController.SendRegisterRouteRequest(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}
