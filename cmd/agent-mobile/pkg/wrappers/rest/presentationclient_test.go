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
	restpresentationclient "github.com/trustbloc/agent-sdk/pkg/controller/rest/presentationclient"
)

func getPresentationClient(t *testing.T) *PresentationClient {
	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetPresentationClient()
	require.NoError(t, err)
	require.NotNil(t, controller)

	pc, ok := controller.(*PresentationClient)
	require.Equal(t, ok, true)

	return pc
}

func TestPresentationClient_SavePresentation(t *testing.T) {
	pc := getPresentationClient(t)

	pc.httpClient = &mockHTTPClient{
		method: http.MethodPost, url: mockAgentURL + restpresentationclient.SavePresentationPath,
	}

	payload, err := json.Marshal(sdscomm.SavePresentationToSDSRequest{})
	require.NoError(t, err)

	resp := pc.SavePresentation(&models.RequestEnvelope{Payload: payload})

	require.NotNil(t, resp)
	require.Nil(t, resp.Error)
	require.Equal(t, []byte{}, resp.Payload)
}
