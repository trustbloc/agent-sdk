/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command //nolint:testpackage // uses internal implementation details

import (
	"fmt"
	"strconv"
	"testing"

	cmdverifiable "github.com/hyperledger/aries-framework-go/pkg/controller/command/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

//nolint:lll
const (
	mockVC = `
{
  "@context":[
     "https://www.w3.org/2018/credentials/v1",
	  "https://trustbloc.github.io/context/vc/examples-v1.jsonld"
  ],
  "id":"http://example.edu/credentials/1989",
  "type":"VerifiableCredential",
  "credentialSubject":{
     "id":"did:example:iuajk1f712ebc6f1c276e12ec21"
  },
  "issuer":{
     "id":"did:example:09s12ec712ebc6f1c671ebfeb1f",
     "name":"Example University"
  },
  "issuanceDate":"2020-01-01T10:54:01Z",
  "credentialStatus":{
     "id":"https://example.gov/status/65",
     "type":"CredentialStatusList2017"
  }
}
`
	mockSignedVC             = `{"@context":["https://www.w3.org/2018/credentials/v1","https://trustbloc.github.io/context/vc/examples-v1.jsonld"],"credentialStatus":{"id":"https://example.gov/status/65","type":"CredentialStatusList2017"},"credentialSubject":"did:example:iuajk1f712ebc6f1c276e12ec21","id":"http://example.edu/credentials/1989","issuanceDate":"2020-01-01T10:54:01Z","issuer":{"id":"did:example:09s12ec712ebc6f1c671ebfeb1f","name":"Example University"},"proof":{"created":"2020-07-13T09:25:45.843216-04:00","jws":"eyJhbGciOiJFZERTQSIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..","proofPurpose":"assertionMethod","type":"Ed25519Signature2018","verificationMethod":"did:peer:123456789abcdefghi#keys-1"},"type":"VerifiableCredential"}`
	mockPresentationResponse = `
{
	"verifiablePresentation": {
		"@context": [
			"https://www.w3.org/2018/credentials/v1"
		],
		"holder": "did:peer:123456789abcdefghi#inbox",
		"proof": {
			"created": "2020-07-10T15:53:25.157489-04:00",
			"jws": "eyJhbGciOiJFZERTQSIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..",
			"proofPurpose": "authentication",
			"type": "Ed25519Signature2018",
			"verificationMethod": "did:peer:123456789abcdefghi#keys-1"
		},
		"type": "VerifiablePresentation",
		"verifiableCredential": [
			{
				"@context": [
					"https://www.w3.org/2018/credentials/v1",
					"https://trustbloc.github.io/context/vc/examples-v1.jsonld"
				],
				"credentialStatus": {
					"id": "https://example.gov/status/65",
					"type": "CredentialStatusList2017"
				},
				"credentialSubject": "did:example:iuajk1f712ebc6f1c276e12ec21",
				"id": "http://example.edu/credentials/1989",
				"issuanceDate": "2020-01-01T10:54:01Z",
				"issuer": {
					"id": "did:example:09s12ec712ebc6f1c671ebfeb1f",
					"name": "Example University"
				},
				"type": "VerifiableCredential"
			}
		]
	}
}`
	mockCredentialName   = "mock_credential" //nolint: gosec // False positive
	mockPresentationName = "mock_vp_name"
	mockCredentialID     = "http://example.edu/credentials/1989" //nolint: gosec // False positive
	mockPresentationID   = "http://example.edu/presentations/1989"
	mockVP               = `{"verifiablePresentation":{"@context":["https://www.w3.org/2018/credentials/v1","https://www.w3.org/2018/credentials/examples/v1"],"type":["VerifiablePresentation"],"id":"http://example.edu/presentations/1989","verifiableCredential":[{"@context":["https://www.w3.org/2018/credentials/v1","https://www.w3.org/2018/credentials/examples/v1"],"credentialSchema":[],"credentialStatus":{"id":"http://issuer.vc.rest.example.com:8070/status/1","type":"CredentialStatusList2017"},"credentialSubject":{"degree":{"degree":"MIT","type":"BachelorDegree"},"id":"did:example:ebfeb1f712ebc6f1c276e12ec21","name":"Jayden Doe","spouse":"did:example:c276e12ec21ebfeb1f712ebc6f1"},"id":"https://example.com/credentials/9315d0fd-da93-436e-9e20-2121f2821df3","issuanceDate":"2020-03-16T22:37:26.544Z","issuer":{"id":"did:elem:EiBJJPdo-ONF0jxqt8mZYEj9Z7FbdC87m2xvN0_HAbcoEg","name":"alice_ca31684e-6cbb-40f9-b7e6-87e1ab5661ae"},"proof":{"created":"2020-04-08T21:19:02Z","jws":"eyJhbGciOiJFZERTQSIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..yGHHYmRp4mWd918SDSzmBDs8eq-SX7WPl8moGB8oJeSqEMmuEiI81D4s5-BPWGmKy3VlCsKJxYrTNqrEGJpNAQ","proofPurpose":"assertionMethod","type":"Ed25519Signature2018","verificationMethod":"did:elem:EiBJJPdo-ONF0jxqt8mZYEj9Z7FbdC87m2xvN0_HAbcoEg#xqc3gS1gz1vch7R3RvNebWMjLvBOY-n_14feCYRPsUo"},"type":["VerifiableCredential","UniversityDegreeCredential"]}]},"name":"sampleVpName"}`
	mockDID              = "did:peer:123456789abcdefghi#inbox"
	emptyJSON            = `{}`
)

func getVerifiableController(t *testing.T) *Verifiable {
	t.Helper()

	a, err := getAgent()
	require.NotNil(t, a)
	require.NoError(t, err)

	vc, err := a.GetVerifiableController()
	require.NoError(t, err)
	require.NotNil(t, vc)

	v, ok := vc.(*Verifiable)
	require.Equal(t, ok, true)

	return v
}

func TestVerifiable_ValidateCredential(t *testing.T) {
	t.Run("test it validates a credential", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := emptyJSON
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.ValidateCredentialCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{"verifiableCredential": %s}`, strconv.Quote(mockVC))

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.ValidateCredential(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_SaveCredential(t *testing.T) {
	t.Run("test it saves a credential", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := emptyJSON
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.SaveCredentialCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{"verifiableCredential": %s, "name": %q}`, strconv.Quote(mockVC), mockCredentialName)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.SaveCredential(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_SavePresentation(t *testing.T) {
	t.Run("test it saves a presentation", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := emptyJSON
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.SavePresentationCommandMethod] = fakeHandler.exec

		payload := mockVP

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.SavePresentation(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_GetCredential(t *testing.T) {
	t.Run("test it gets a credential", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := fmt.Sprintf(`{"verifiableCredential": %s}`, strconv.Quote(mockVC))
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.GetCredentialCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{"id":%q}`, mockCredentialID)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.GetCredential(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_SignCredential(t *testing.T) {
	t.Run("test it signs a credential", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := fmt.Sprintf(`{"verifiableCredential": %s}`, strconv.Quote(mockSignedVC))
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.SignCredentialCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{
		"credential": %s,
		"did": "%s",
		"signatureType": "%s"
}`, strconv.Quote(mockVC), mockDID, cmdverifiable.Ed25519Signature2018)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.SignCredential(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_GetPresentation(t *testing.T) {
	t.Run("test it gets a presentation", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := fmt.Sprintf(`{"verifiablePresentation": %s}`, strconv.Quote(mockVP))
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.GetPresentationCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{"id":%q}`, mockPresentationID)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.GetPresentation(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_GetCredentialByName(t *testing.T) {
	t.Run("test it gets a verifiable credential by name", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := fmt.Sprintf(`{"name": %s, "id": %s}`, mockCredentialName, mockCredentialID)
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.GetCredentialByNameCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{"name":%q}`, mockCredentialName)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.GetCredentialByName(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_GetCredentials(t *testing.T) {
	t.Run("test it gets all stored verifiable credentials", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := fmt.Sprintf(`{"result": [{"name": %q, "id": %q}, {"name": %q, "id": %q"}]`,
			mockCredentialName, mockCredentialID, mockCredentialName, mockCredentialID)
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.GetCredentialsCommandMethod] = fakeHandler.exec

		payload := "{}"

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.GetCredentials(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_GetPresentations(t *testing.T) {
	t.Run("test it gets all stored verifiable presentations", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := fmt.Sprintf(`{"result": [{"name": %q, "id": %q}, {"name": %q, "id": %q"}]`,
			mockPresentationName, mockPresentationID, mockPresentationName, mockPresentationID)
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.GetPresentationsCommandMethod] = fakeHandler.exec

		payload := "{}"

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.GetPresentations(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_GeneratePresentation(t *testing.T) {
	t.Run("test it generates a presentation", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := mockPresentationResponse
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.GeneratePresentationCommandMethod] = fakeHandler.exec

		credList := fmt.Sprintf(`[%s, %s]`, mockVC, mockVC)
		payload := fmt.Sprintf(`{
		"verifiableCredential": %s,
		"did": "%s",
		"signatureType": "%s"
}`, credList, mockDID, cmdverifiable.Ed25519Signature2018)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.GeneratePresentation(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_GeneratePresentationByID(t *testing.T) {
	t.Run("test it generates a presentation by id", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := mockPresentationResponse
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.GeneratePresentationByIDCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{
		"id": %q,
		"did": "%s",
		"signatureType": "%s"
}`, mockCredentialID, mockDID, cmdverifiable.Ed25519Signature2018)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.GeneratePresentationByID(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_RemoveCredentialByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := ``
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.RemoveCredentialByNameCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{"name":%q}`, mockCredentialName)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.RemoveCredentialByName(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}

func TestVerifiable_RemovePresentationByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		v := getVerifiableController(t)

		mockResponse := ``
		fakeHandler := mockCommandRunner{data: []byte(mockResponse)}
		v.handlers[cmdverifiable.RemovePresentationByNameCommandMethod] = fakeHandler.exec

		payload := fmt.Sprintf(`{"name":%q}`, mockPresentationName)

		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := v.RemovePresentationByName(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t,
			mockResponse,
			string(resp.Payload))
	})
}
