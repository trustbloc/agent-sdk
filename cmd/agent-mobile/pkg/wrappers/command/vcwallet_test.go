/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command // nolint:testpackage // uses internal implementation details

import (
	_ "embed" //nolint:gci // required for go:embed
	"encoding/json"
	"fmt"
	"testing"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/config"
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

	cmdvcwallet "github.com/hyperledger/aries-framework-go/pkg/controller/command/vcwallet"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ldcontext"
	mockldstore "github.com/hyperledger/aries-framework-go/pkg/mock/ld"
	ldstore "github.com/hyperledger/aries-framework-go/pkg/store/ld"
	"github.com/stretchr/testify/require"
)

const (
	sampleUserAuth = `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`
	sampleUDCVC    = `{
      "@context": [
        "https://www.w3.org/2018/credentials/v1",
        "https://www.w3.org/2018/credentials/examples/v1"
      ],
     "credentialSchema": [],
      "credentialSubject": {
        "degree": {
          "type": "BachelorDegree",
          "university": "MIT"
        },
        "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
        "name": "Jayden Doe",
        "spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
      },
      "expirationDate": "2020-01-01T19:23:24Z",
      "id": "http://example.edu/credentials/1877",
      "issuanceDate": "2010-01-01T19:23:24Z",
      "issuer": {
        "id": "did:example:76e12ec712ebc6f1c221ebfeb1f",
        "name": "Example University"
      },
      "referenceNumber": 83294847,
      "type": [
        "VerifiableCredential",
        "UniversityDegreeCredential"
      ]
    }`
	sampleUDCVC2 = `{
		"@context": [
		  "https://www.w3.org/2018/credentials/v1",
		  "https://www.w3.org/2018/credentials/examples/v1"
		],
	   "credentialSchema": [],
		"credentialSubject": {
		  "degree": {
			"type": "BachelorDegree",
			"university": "MIT"
		  },
		  "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
		  "name": "Jayden Doe",
		  "spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
		},
		"expirationDate": "2020-01-01T19:23:24Z",
		"id": "http://example.edu/credentials/1888",
		"issuanceDate": "2010-01-01T19:23:24Z",
		"issuer": {
		  "id": "did:example:76e12ec712ebc6f1c221ebfeb1f",
		  "name": "Example University"
		},
		"referenceNumber": 83294847,
		"type": [
		  "VerifiableCredential",
		  "UniversityDegreeCredential"
		]
	  }`
	sampleQueryByExample = `{
                        "reason": "Please present your identity document.",
                        "example": {
                            "@context": [
								"https://www.w3.org/2018/credentials/v1",
								"https://www.w3.org/2018/credentials/examples/v1"
                            ],
                            "type": ["UniversityDegreeCredential"],
							"trustedIssuer": [
              					{
                					"issuer": "urn:some:required:issuer"
              					},
								{
                					"required": true,
                					"issuer": "did:example:76e12ec712ebc6f1c221ebfeb1f"
              					}
							],
							"credentialSubject": {
								"id": "did:example:ebfeb1f712ebc6f1c276e12ec21"	
							}
                        }
                	}`
	sampleQueryByFrame = `{
                    "reason": "Please provide your Passport details.",
                    "frame": {
                        "@context": [
                            "https://www.w3.org/2018/credentials/v1",
                            "https://w3id.org/citizenship/v1",
                            "https://w3id.org/security/bbs/v1"
                        ],
                        "type": ["VerifiableCredential", "PermanentResidentCard"],
                        "@explicit": true,
                        "identifier": {},
                        "issuer": {},
                        "issuanceDate": {},
                        "credentialSubject": {
                            "@explicit": true,
                            "name": {},
                            "spouse": {}
                        }
                    },
                    "trustedIssuer": [
                        {
                            "issuer": "did:example:76e12ec712ebc6f1c221ebfeb1f",
                            "required": true
                        }
                    ],
                    "required": true
                }`
)

// nolint:gochecknoglobals // embedded test contexts
var (
	//go:embed testdata/contexts/credentials-examples_v1.jsonld
	credentialExamples []byte
	//go:embed testdata/contexts/examples_v1.jsonld
	vcExamples []byte
	//go:embed testdata/contexts/odrl.jsonld
	odrl []byte
	//go:embed testdata/contexts/citizenship_v1.jsonld
	citizenship []byte
	//go:embed testdata/contexts/governance.jsonld
	governance []byte
	//go:embed testdata/contexts/lds-jws2020-v1.jsonld
	jws2020 []byte
)

// DocumentLoader returns a document loader with preloaded test contexts.
func DocumentLoader(t *testing.T) *ld.DocumentLoader {
	t.Helper()

	ldStore := &mockLDStoreProvider{
		ContextStore:        mockldstore.NewMockContextStore(),
		RemoteProviderStore: mockldstore.NewMockRemoteProviderStore(),
	}

	loader, err := ld.NewDocumentLoader(ldStore,
		ld.WithExtraContexts(
			ldcontext.Document{
				URL:     "https://www.w3.org/2018/credentials/examples/v1",
				Content: credentialExamples,
			},
			ldcontext.Document{
				URL:     "https://trustbloc.github.io/context/vc/examples-v1.jsonld",
				Content: vcExamples,
			},
			ldcontext.Document{
				URL:     "https://www.w3.org/ns/odrl.jsonld",
				Content: odrl,
			},
			ldcontext.Document{
				URL:         "https://w3id.org/citizenship/v1",
				DocumentURL: "https://w3c-ccg.github.io/citizenship-vocab/contexts/citizenship-v1.jsonld",
				Content:     citizenship,
			},
			ldcontext.Document{
				URL:     "https://trustbloc.github.io/context/governance/context.jsonld",
				Content: governance,
			},
			ldcontext.Document{
				URL:     "https://w3c-ccg.github.io/lds-jws2020/contexts/lds-jws2020-v1.json",
				Content: jws2020,
			},
		),
	)
	require.NoError(t, err)

	return loader
}

type mockLDStoreProvider struct {
	ContextStore        ldstore.ContextStore
	RemoteProviderStore ldstore.RemoteProviderStore
}

func (p *mockLDStoreProvider) JSONLDContextStore() ldstore.ContextStore {
	return p.ContextStore
}

func (p *mockLDStoreProvider) JSONLDRemoteProviderStore() ldstore.RemoteProviderStore {
	return p.RemoteProviderStore
}

func getVCWalletController(t *testing.T) *VCWallet {
	t.Helper()

	a, err := getAgentWithOpts(&config.Options{DocumentLoader: DocumentLoader(t)})
	require.NotNil(t, a)
	require.NoError(t, err)

	controller, err := a.GetVCWalletController()
	require.NoError(t, err)
	require.NotNil(t, controller)

	v, ok := controller.(*VCWallet)
	require.Equal(t, ok, true)

	return v
}

func TestVCWallet_CreateProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		vcwalletController := getVCWalletController(t)
		require.NotNil(t, vcwalletController)

		createProfilePayload := sampleUserAuth

		createProfileReq := &models.RequestEnvelope{Payload: []byte(createProfilePayload)}

		createProfileResp := vcwalletController.CreateProfile(createProfileReq)
		require.NotNil(t, createProfileResp)
		require.Nil(t, createProfileResp.Error)
		require.Equal(t,
			``,
			string(createProfileResp.Payload))
	})
}

func TestVCWallet_ProfileExists(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		vcwalletController := getVCWalletController(t)
		require.NotNil(t, vcwalletController)

		createProfilePayload := sampleUserAuth

		createProfileReq := &models.RequestEnvelope{Payload: []byte(createProfilePayload)}

		createProfileResp := vcwalletController.CreateProfile(createProfileReq)
		require.NotNil(t, createProfileResp)
		require.Nil(t, createProfileResp.Error)
		require.Equal(t,
			``,
			string(createProfileResp.Payload))

		// check that profile exists, user1 should exist
		payloadExists := `{"userID":"user1"}`
		reqExists := &models.RequestEnvelope{Payload: []byte(payloadExists)}

		respExists := vcwalletController.ProfileExists(reqExists)
		require.NotNil(t, respExists)
		require.Nil(t, respExists.Error)
		require.Equal(t,
			``,
			string(respExists.Payload))

		// check that profile exists, this should not exist so Error is not nil
		payloadNotExists := `{"userID":"user12"}`
		reqNotExists := &models.RequestEnvelope{Payload: []byte(payloadNotExists)}

		respNotExists := vcwalletController.ProfileExists(reqNotExists)
		require.NotNil(t, respNotExists)
		require.NotNil(t, respNotExists.Error)
		require.Equal(t, &models.CommandError{Message: "profile does not exist", Code: 12015, Type: 1}, respNotExists.Error)
	})
}

func TestVCWallet_Open_Close(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		vcwalletController := getVCWalletController(t)
		require.NotNil(t, vcwalletController)
		openPayload := sampleUserAuth

		openReq := &models.RequestEnvelope{Payload: []byte(openPayload)}

		// should fail, user doesn't have a wallet yet
		openResp := vcwalletController.Open(openReq)
		require.NotNil(t, openResp)
		require.NotNil(t, openResp.Error)

		createProfilePayload := sampleUserAuth
		createProfileReq := &models.RequestEnvelope{Payload: []byte(createProfilePayload)}

		createProfileResp := vcwalletController.CreateProfile(createProfileReq)
		require.NotNil(t, createProfileResp)
		require.Nil(t, createProfileResp.Error)
		require.Equal(t,
			``,
			string(createProfileResp.Payload))

		// now open should succeed
		openResp = vcwalletController.Open(openReq)
		require.NotNil(t, openResp)
		require.Nil(t, openResp.Error)

		var tokenResponse cmdvcwallet.UnlockWalletResponse
		if err := json.Unmarshal(openResp.Payload, &tokenResponse); err != nil {
			t.Fail()
		} else {
			require.NotNil(t,
				tokenResponse.Token)
			require.NotEqual(t,
				``,
				tokenResponse.Token)
		}

		closePayload := `{"userID":"user1"}`

		closeReq := &models.RequestEnvelope{Payload: []byte(closePayload)}
		closeResp := vcwalletController.Close(closeReq)
		require.NotNil(t, closeResp)
		require.Nil(t, closeResp.Error)

		var lockResponse cmdvcwallet.LockWalletResponse
		if err := json.Unmarshal(closeResp.Payload, &lockResponse); err != nil {
			// fail
			t.Fail()
		} else {
			require.Equal(t,
				lockResponse.Closed,
				true)
		}

		// close again, should return closed = false
		closeResp = vcwalletController.Close(closeReq)
		require.NotNil(t, closeResp)
		require.Nil(t, closeResp.Error)

		if err := json.Unmarshal(closeResp.Payload, &lockResponse); err != nil {
			// fail
			t.Fail()
		} else {
			require.Equal(t,
				lockResponse.Closed,
				false)
		}
	})
}

// nolint: lll
func TestVCWallet_Add_Get_GetAll(t *testing.T) {
	vcwalletController := getVCWalletController(t)
	require.NotNil(t, vcwalletController)

	var tokenResponse cmdvcwallet.UnlockWalletResponse

	t.Run("create profile", func(t *testing.T) {
		// create profile
		createProfilePayload := sampleUserAuth
		createProfileReq := &models.RequestEnvelope{Payload: []byte(createProfilePayload)}
		createProfileResp := vcwalletController.CreateProfile(createProfileReq)
		require.NotNil(t, createProfileResp)
		require.Nil(t, createProfileResp.Error)
		require.Equal(t,
			``,
			string(createProfileResp.Payload))
	})

	t.Run("unlock", func(t *testing.T) {
		// open the wallet
		openPayload := sampleUserAuth
		openReq := &models.RequestEnvelope{Payload: []byte(openPayload)}

		openResp := vcwalletController.Open(openReq)
		require.NotNil(t, openResp)
		require.Nil(t, openResp.Error)

		if err := json.Unmarshal(openResp.Payload, &tokenResponse); err != nil {
			t.Fail()
		} else {
			require.NotNil(t,
				tokenResponse.Token)
			require.NotEqual(t,
				``,
				tokenResponse.Token)
		}
	})

	t.Run("add credential", func(t *testing.T) {
		// add proper content
		addPayload := fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType":"credential", "content":%s}`, tokenResponse.Token, sampleUDCVC)
		fmt.Println(addPayload)
		addReq := &models.RequestEnvelope{Payload: []byte(addPayload)}

		addResp := vcwalletController.Add(addReq)
		require.NotNil(t, addResp)
		require.Nil(t, addResp.Error)
		require.Equal(t,
			``,
			string(addResp.Payload))
	})

	t.Run("get content", func(t *testing.T) {
		// get it back
		getPayload := fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType": "credential", "contentID": "http://example.edu/credentials/1877"}`, tokenResponse.Token)
		getReq := &models.RequestEnvelope{Payload: []byte(getPayload)}
		getResp := vcwalletController.Get(getReq)
		require.NotNil(t, getResp)
		require.Nil(t, getResp.Error)

		var getContentResponse cmdvcwallet.GetContentResponse
		if err := json.Unmarshal(getResp.Payload, &getContentResponse); err != nil {
			t.Fail()
		} else {
			require.NotEmpty(t, getContentResponse.Content)
		}
	})

	t.Run("get all", func(t *testing.T) {
		addPayload := fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType":"credential", "content":%s}`, tokenResponse.Token, sampleUDCVC2)
		fmt.Println(addPayload)
		addReq := &models.RequestEnvelope{Payload: []byte(addPayload)}
		addResp := vcwalletController.Add(addReq)
		require.NotNil(t, addResp)
		require.Nil(t, addResp.Error)
		require.Equal(t,
			``,
			string(addResp.Payload))

		// get all
		getPayload := fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType": "credential"}`, tokenResponse.Token)
		getReq := &models.RequestEnvelope{Payload: []byte(getPayload)}
		getResp := vcwalletController.GetAll(getReq)
		require.NotNil(t, getResp)
		require.Nil(t, getResp.Error)

		var getAllContentResponse cmdvcwallet.GetAllContentResponse
		if err := json.Unmarshal(getResp.Payload, &getAllContentResponse); err != nil {
			t.Fail()
		} else {
			require.NotEmpty(t, getAllContentResponse.Contents)
			require.Len(t, getAllContentResponse.Contents, 2)
		}
	})

	t.Run("remove", func(t *testing.T) {
		// remove one
		removePayload := fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType": "credential", "contentID": "http://example.edu/credentials/1877"}`, tokenResponse.Token)
		removeReq := &models.RequestEnvelope{Payload: []byte(removePayload)}
		removeResp := vcwalletController.Remove(removeReq)
		require.NotNil(t, removeResp)
		require.Nil(t, removeResp.Error)
		require.Equal(t,
			``,
			string(removeResp.Payload))
	})

	t.Run("query", func(t *testing.T) {
		payload := fmt.Sprintf(`{"userID":"user1", "auth": "%s", "query": [
			{"type":"QueryByExample", "credentialQuery":[%s]},
			{"type":"QueryByFrame", "credentialQuery":[%s]}
		] }`, tokenResponse.Token, sampleQueryByExample, sampleQueryByFrame)
		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := vcwalletController.Query(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)

		var response map[string]interface{}
		if err := json.Unmarshal(resp.Payload, &response); err != nil {
			t.Fail()
		} else {
			require.NotEmpty(t, response["results"])
		}
	})

	t.Run("query with invalid user", func(t *testing.T) {
		payload := fmt.Sprintf(`{"userID":"user12", "auth": "%s", "query": [
			{"type":"QueryByExample", "credentialQuery":[%s]},
			{"type":"QueryByFrame", "credentialQuery":[%s]}
		] }`, tokenResponse.Token, sampleQueryByExample, sampleQueryByFrame)
		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := vcwalletController.Query(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
	})

	t.Run("query with invalid auth", func(t *testing.T) {
		payload := fmt.Sprintf(`{"userID":"user1", "auth": "%s", "query": [
			{"type":"QueryByExample", "credentialQuery":[%s]},
			{"type":"QueryByFrame", "credentialQuery":[%s]}
		] }`, "crap", sampleQueryByExample, sampleQueryByFrame)
		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := vcwalletController.Query(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
	})

	t.Run("query with invalid query", func(t *testing.T) {
		payload := fmt.Sprintf(`{"userID":"user12", "auth": "%s", "query": [
			{"type":"QueryByXExample", "credentialQuery":[%s]},
			{"type":"QueryByXFrame", "credentialQuery":[%s]}
		] }`, tokenResponse.Token, sampleQueryByExample, sampleQueryByFrame)
		req := &models.RequestEnvelope{Payload: []byte(payload)}
		resp := vcwalletController.Query(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
	})
}
