/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest //nolint:testpackage // uses internal implementation details

import (
	"net/http"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/controller/rest/kms"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

func getKMSController(t *testing.T) *KMS {
	t.Helper()

	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetKMSController()
	require.NoError(t, err)
	require.NotNil(t, controller)

	k, ok := controller.(*KMS)
	require.Equal(t, ok, true)

	return k
}

func TestKMS_CreateKeySet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		controller := getKMSController(t)

		reqData := `{"keyType":"ED25519"}`
		mockResponse := `{"keyID":"keyID","publicKey":"cHVibGljS2V5"}`

		controller.httpClient = &mockHTTPClient{
			data:   mockResponse,
			method: http.MethodPost, url: mockAgentURL + kms.CreateKeySetPath,
		}

		req := &models.RequestEnvelope{Payload: []byte(reqData)}
		resp := controller.CreateKeySet(req)

		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t, mockResponse, string(resp.Payload))
	})
}

func TestKMS_ImportKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		controller := getKMSController(t)

		reqData := `{"kty":"OKP","kid":"kid","crv":"Ed25519","alg":"EdDSA",
"x":"jXAvdkE8oHbFat1HYkdq3FXsuPdGtdl8NhKr163kikA","d":"QlXTAvl0V7Kh7ckWXTVmdAdZZQcIdZ0yqXxwvw9QX04"}`
		mockResponse := emptyJSON

		controller.httpClient = &mockHTTPClient{
			data:   mockResponse,
			method: http.MethodPost, url: mockAgentURL + kms.ImportKeyPath,
		}

		req := &models.RequestEnvelope{Payload: []byte(reqData)}
		resp := controller.ImportKey(req)

		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t, mockResponse, string(resp.Payload))
	})
}
