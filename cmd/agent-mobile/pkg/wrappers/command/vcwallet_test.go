/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command // nolint:testpackage // uses internal implementation details

import (
	"fmt"

	"encoding/json"
	"testing"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

	cmdvcwallet "github.com/hyperledger/aries-framework-go/pkg/controller/command/vcwallet"

	"github.com/stretchr/testify/require"
)

const (
	sampleUserID           = "sample-user01"
	samplePassPhrase       = "fakepassphrase"
	sampleKeyStoreURL      = "sample/keyserver/test"
	sampleEDVServerURL     = "sample-edv-url"
	sampleEDVVaultID       = "sample-edv-vault-id"
	sampleEDVEncryptionKID = "sample-edv-encryption-kid"
	sampleEDVMacKID        = "sample-edv-mac-kid"
	sampleCommandError     = "sample-command-error-01"
	sampleFakeTkn          = "sample-fake-token-01"
	sampleFakeCapability   = "sample-fake-capability-01"
	sampleDIDKey           = "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
	sampleUDCVC            = `{
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
	sampleMetadata = `{
        "@context": ["https://w3id.org/wallet/v1"],
        "id": "urn:uuid:2905324a-9524-11ea-bb37-0242ac130002",
        "type": "Metadata",
        "name": "Ropsten Testnet HD Accounts",
        "image": "https://via.placeholder.com/150",
        "description": "My Ethereum TestNet Accounts",
        "tags": ["professional", "organization"],
        "correlation": ["urn:uuid:4058a72a-9523-11ea-bb37-0242ac130002"],
        "hdPath": "m’/44’/60’/0’",
        "target": ["urn:uuid:c410e44a-9525-11ea-bb37-0242ac130002"]
    }`
	sampleBBSVC = `{
            "@context": ["https://www.w3.org/2018/credentials/v1", "https://www.w3.org/2018/credentials/examples/v1", "https://w3id.org/security/bbs/v1"],
            "credentialSubject": {
                "degree": {"type": "BachelorDegree", "university": "MIT"},
                "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
                "name": "Jayden Doe",
                "spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
            },
            "expirationDate": "2020-01-01T19:23:24Z",
            "id": "http://example.edu/credentials/1872",
            "issuanceDate": "2010-01-01T19:23:24Z",
            "issuer": {"id": "did:example:76e12ec712ebc6f1c221ebfeb1f", "name": "Example University"},
            "proof": {
                "created": "2021-03-29T13:27:36.483097-04:00",
                "proofPurpose": "assertionMethod",
                "proofValue": "rw7FeV6K1wimnYogF9qd-N0zmq5QlaIoszg64HciTca-mK_WU4E1jIusKTT6EnN2GZz04NVPBIw4yhc0kTwIZ07etMvfWUlHt_KMoy2CfTw8FBhrf66q4h7Qcqxh_Kxp6yCHyB4A-MmURlKKb8o-4w",
                "type": "BbsBlsSignature2020",
                "verificationMethod": "did:key:zUC72c7u4BYVmfYinDceXkNAwzPEyuEE23kUmJDjLy8495KH3pjLwFhae1Fww9qxxRdLnS2VNNwni6W3KbYZKsicDtiNNEp76fYWR6HCD8jAz6ihwmLRjcHH6kB294Xfg1SL1qQ#zUC72c7u4BYVmfYinDceXkNAwzPEyuEE23kUmJDjLy8495KH3pjLwFhae1Fww9qxxRdLnS2VNNwni6W3KbYZKsicDtiNNEp76fYWR6HCD8jAz6ihwmLRjcHH6kB294Xfg1SL1qQ"
            },
            "referenceNumber": 83294847,
            "type": ["VerifiableCredential", "UniversityDegreeCredential"]
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
	sampleFrame = `
		{
			"@context": [
				"https://www.w3.org/2018/credentials/v1",
				"https://www.w3.org/2018/credentials/examples/v1",
				"https://w3id.org/security/bbs/v1"
			],
  			"type": ["VerifiableCredential", "UniversityDegreeCredential"],
  			"@explicit": true,
  			"identifier": {},
  			"issuer": {},
  			"issuanceDate": {},
  			"credentialSubject": {
    			"@explicit": true,
    			"degree": {},
    			"name": {}
  			}
		}
	`
	sampleKeyContentBase58 = `{
  			"@context": ["https://w3id.org/wallet/v1"],
  		  	"id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
  		  	"controller": "did:example:123456789abcdefghi",
			"type": "Ed25519VerificationKey2018",
			"privateKeyBase58":"2MP5gWCnf67jvW3E4Lz8PpVrDWAXMYY1sDxjnkEnKhkkbKD7yP2mkVeyVpu5nAtr3TeDgMNjBPirk2XcQacs3dvZ"
  		}`
	sampleDIDResolutionResponse = `{
        "@context": [
            "https://w3id.org/wallet/v1",
            "https://w3id.org/did-resolution/v1"
        ],
        "id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
        "type": ["DIDResolutionResponse"],
        "name": "Farming Sensor DID Document",
        "image": "https://via.placeholder.com/150",
        "description": "An IoT device in the middle of a corn field.",
        "tags": ["professional"],
        "correlation": ["4058a72a-9523-11ea-bb37-0242ac130002"],
        "created": "2017-06-18T21:19:10Z",
        "expires": "2026-06-18T21:19:10Z",
        "didDocument": {
            "@context": [
                "https://w3id.org/did/v0.11"
            ],
            "id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
            "publicKey": [
                {
                    "id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
                    "type": "Ed25519VerificationKey2018",
                    "controller": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
                    "publicKeyBase58": "8jkuMBqmu1TRA6is7TT5tKBksTZamrLhaXrg9NAczqeh"
                }
            ],
            "authentication": [
                "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
            ],
            "assertionMethod": [
                "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
            ],
            "capabilityDelegation": [
                "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
            ],
            "capabilityInvocation": [
                "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
            ],
            "keyAgreement": [
                {
                    "id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6LSmjNfS5FC9W59JtPZq7fHgrjThxsidjEhZeMxCarbR998",
                    "type": "X25519KeyAgreementKey2019",
                    "controller": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
                    "publicKeyBase58": "B4CVumSL43MQDW1oJU9LNGWyrpLbw84YgfeGi8D4hmNN"
                }
            ]
        }
    }`
)

func getVCWalletController(t *testing.T) *VCWallet {
	t.Helper()

	a, err := getAgentWithOpts(t)
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

		createProfilePayload := `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`

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

		createProfilePayload := `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`

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
		openPayload := `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`

		openReq := &models.RequestEnvelope{Payload: []byte(openPayload)}

		// should fail, user doesn't have a wallet yet
		var openResp = vcwalletController.Open(openReq)
		require.NotNil(t, openResp)
		require.NotNil(t, openResp.Error)

		createProfilePayload := `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`
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
		if err := json.Unmarshal([]byte(openResp.Payload), &tokenResponse); err != nil {
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
		var closeResp = vcwalletController.Close(closeReq)
		require.NotNil(t, closeResp)
		require.Nil(t, closeResp.Error)

		var lockResponse cmdvcwallet.LockWalletResponse
		if err := json.Unmarshal([]byte(closeResp.Payload), &lockResponse); err != nil {
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

		if err := json.Unmarshal([]byte(closeResp.Payload), &lockResponse); err != nil {
			// fail
			t.Fail()
		} else {
			require.Equal(t,
				lockResponse.Closed,
				false)
		}
	})
}

func TestVCWallet_Add_Get_GetAll(t *testing.T) {
	vcwalletController := getVCWalletController(t)
	require.NotNil(t, vcwalletController)
	var tokenResponse cmdvcwallet.UnlockWalletResponse

	t.Run("create profile", func(t *testing.T) {

		// create profile
		createProfilePayload := `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`
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
		openPayload := `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`
		openReq := &models.RequestEnvelope{Payload: []byte(openPayload)}

		var openResp = vcwalletController.Open(openReq)
		require.NotNil(t, openResp)
		require.Nil(t, openResp.Error)

		if err := json.Unmarshal([]byte(openResp.Payload), &tokenResponse); err != nil {
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
		var addPayload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType":"credential", "content":%s}`, tokenResponse.Token, sampleUDCVC)
		fmt.Println(addPayload)
		var addReq = &models.RequestEnvelope{Payload: []byte(addPayload)}
		var addResp = vcwalletController.Add(addReq)
		require.NotNil(t, addResp)
		require.Nil(t, addResp.Error)
		require.Equal(t,
			``,
			string(addResp.Payload))
	})

	t.Run("get content", func(t *testing.T) {

		// get it back
		var getPayload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType": "credential", "contentID": "http://example.edu/credentials/1877"}`, tokenResponse.Token)
		var getReq = &models.RequestEnvelope{Payload: []byte(getPayload)}
		var getResp = vcwalletController.Get(getReq)
		require.NotNil(t, getResp)
		require.Nil(t, getResp.Error)

		var getContentResponse cmdvcwallet.GetContentResponse
		if err := json.Unmarshal([]byte(getResp.Payload), &getContentResponse); err != nil {
			t.Fail()
		} else {
			require.NotEmpty(t, getContentResponse.Content)
		}
	})

	t.Run("get all", func(t *testing.T) {
		var addPayload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType":"credential", "content":%s}`, tokenResponse.Token, sampleUDCVC2)
		fmt.Println(addPayload)
		var addReq = &models.RequestEnvelope{Payload: []byte(addPayload)}
		var addResp = vcwalletController.Add(addReq)
		require.NotNil(t, addResp)
		require.Nil(t, addResp.Error)
		require.Equal(t,
			``,
			string(addResp.Payload))

		// get all
		var getPayload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType": "credential"}`, tokenResponse.Token)
		var getReq = &models.RequestEnvelope{Payload: []byte(getPayload)}
		var getResp = vcwalletController.GetAll(getReq)
		require.NotNil(t, getResp)
		require.Nil(t, getResp.Error)

		var getAllContentResponse cmdvcwallet.GetAllContentResponse
		if err := json.Unmarshal([]byte(getResp.Payload), &getAllContentResponse); err != nil {
			t.Fail()
		} else {
			require.NotEmpty(t, getAllContentResponse.Contents)
			require.Len(t, getAllContentResponse.Contents, 2)
		}
	})

	t.Run("remove", func(t *testing.T) {
		// remove one
		var removePayload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType": "credential", "contentID": "http://example.edu/credentials/1877"}`, tokenResponse.Token)
		var removeReq = &models.RequestEnvelope{Payload: []byte(removePayload)}
		var removeResp = vcwalletController.Remove(removeReq)
		require.NotNil(t, removeResp)
		require.Nil(t, removeResp.Error)
		require.Equal(t,
			``,
			string(removeResp.Payload))

	})

	t.Run("query", func(t *testing.T) {
		var payload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "query": [
			{"type":"QueryByExample", "credentialQuery":[%s]},
			{"type":"QueryByFrame", "credentialQuery":[%s]}
		] }`, tokenResponse.Token, sampleQueryByExample, sampleQueryByFrame)
		var req = &models.RequestEnvelope{Payload: []byte(payload)}
		var resp = vcwalletController.Query(req)
		require.NotNil(t, resp)
		require.Nil(t, resp.Error)

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(resp.Payload), &response); err != nil {
			t.Fail()
		} else {
			require.NotEmpty(t, response["results"])
		}
	})

	t.Run("query with invalid user", func(t *testing.T) {

		var payload = fmt.Sprintf(`{"userID":"user12", "auth": "%s", "query": [
			{"type":"QueryByExample", "credentialQuery":[%s]},
			{"type":"QueryByFrame", "credentialQuery":[%s]}
		] }`, tokenResponse.Token, sampleQueryByExample, sampleQueryByFrame)
		var req = &models.RequestEnvelope{Payload: []byte(payload)}
		var resp = vcwalletController.Query(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)

	})

	t.Run("query with invalid auth", func(t *testing.T) {

		var payload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "query": [
			{"type":"QueryByExample", "credentialQuery":[%s]},
			{"type":"QueryByFrame", "credentialQuery":[%s]}
		] }`, "crap", sampleQueryByExample, sampleQueryByFrame)
		var req = &models.RequestEnvelope{Payload: []byte(payload)}
		var resp = vcwalletController.Query(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)

	})

	t.Run("query with invalid query", func(t *testing.T) {

		var payload = fmt.Sprintf(`{"userID":"user12", "auth": "%s", "query": [
			{"type":"QueryByXExample", "credentialQuery":[%s]},
			{"type":"QueryByXFrame", "credentialQuery":[%s]}
		] }`, tokenResponse.Token, sampleQueryByExample, sampleQueryByFrame)
		var req = &models.RequestEnvelope{Payload: []byte(payload)}
		var resp = vcwalletController.Query(req)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)

	})

}

func TestVCWallet_Issue(t *testing.T) {
	vcwalletController := getVCWalletController(t)
	require.NotNil(t, vcwalletController)
	var tokenResponse cmdvcwallet.UnlockWalletResponse

	// create profile
	createProfilePayload := `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`
	createProfileReq := &models.RequestEnvelope{Payload: []byte(createProfilePayload)}
	createProfileResp := vcwalletController.CreateProfile(createProfileReq)
	require.NotNil(t, createProfileResp)
	require.Nil(t, createProfileResp.Error)
	require.Equal(t,
		``,
		string(createProfileResp.Payload))

	// open the wallet
	openPayload := `{"userID":"user1", "localKMSPassphrase": "fakepassphrase"}`
	openReq := &models.RequestEnvelope{Payload: []byte(openPayload)}

	var openResp = vcwalletController.Open(openReq)
	require.NotNil(t, openResp)
	require.Nil(t, openResp.Error)

	if err := json.Unmarshal([]byte(openResp.Payload), &tokenResponse); err != nil {
		t.Fail()
	} else {
		require.NotNil(t,
			tokenResponse.Token)
		require.NotEqual(t,
			``,
			tokenResponse.Token)
	}

	// add proper content
	var addPayload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType":"key", "content":%s}`, tokenResponse.Token, sampleKeyContentBase58)
	addContent(t, vcwalletController, addPayload)

	addPayload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "contentType":"didResolutionResponse", "content":%s}`, tokenResponse.Token, sampleDIDResolutionResponse)
	addContent(t, vcwalletController, addPayload)

	// t.Run("issue credential", func(t *testing.T) {
	// 	var payload = fmt.Sprintf(`{"userID":"user1", "auth": "%s", "credential":%s, "proofOptions":{"controller":"%s"}}`, tokenResponse.Token, sampleUDCVC, sampleDIDKey)
	// 	var req = &models.RequestEnvelope{Payload: []byte(payload)}
	// 	var resp = vcwalletController.Issue(req)
	// 	require.NotNil(t, resp)
	// 	require.Nil(t, resp.Error)

	// 	var issueResponse cmdvcwallet.IssueResponse
	// 	if err := json.Unmarshal([]byte(openResp.Payload), &issueResponse); err != nil {
	// 		t.Fail()
	// 	} else {
	// 		// vc, isueerr := verifiable.ParseCredential(issueResponse.Credential, verifiable.WithDisabledProofCheck())
	// 		// fmt.Println("%s", issueResponse.Credential)
	// 		require.Len(t, issueResponse.Credential.Proofs, 1)
	// 	}

	// })
}

func addContent(t *testing.T, vcwalletController *VCWallet, addPayload string) {
	var addReq = &models.RequestEnvelope{Payload: []byte(addPayload)}
	var addResp = vcwalletController.Add(addReq)
	require.NotNil(t, addResp)
	require.Nil(t, addResp.Error)
}
