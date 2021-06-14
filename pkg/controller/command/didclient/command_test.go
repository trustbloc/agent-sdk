/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package didclient // nolint:testpackage // uses internal implementation details

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/crypto/primitive/bbs12381g2pub"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/sidetree/doc"
	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	mockprotocol "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockkms "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	mockvdr "github.com/hyperledger/aries-framework-go/pkg/mock/vdr"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
)

//nolint:lll
const sampleDoc = `{
  "@context": ["https://www.w3.org/ns/did/v1","https://w3id.org/did/v2"],
  "id": "did:peer:21tDAKCERh95uGgKbJNHYp",
  "verificationMethod": [
    {
      "id": "did:peer:123456789abcdefghi#keys-1",
      "type": "Secp256k1VerificationKey2018",
      "controller": "did:peer:123456789abcdefghi",
      "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
    },
    {
      "id": "did:peer:123456789abcdefghw#key2",
      "type": "RsaVerificationKey2018",
      "controller": "did:peer:123456789abcdefghw",
      "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAryQICCl6NZ5gDKrnSztO\n3Hy8PEUcuyvg/ikC+VcIo2SFFSf18a3IMYldIugqqqZCs4/4uVW3sbdLs/6PfgdX\n7O9D22ZiFWHPYA2k2N744MNiCD1UE+tJyllUhSblK48bn+v1oZHCM0nYQ2NqUkvS\nj+hwUU3RiWl7x3D2s9wSdNt7XUtW05a/FXehsPSiJfKvHJJnGOX0BgTvkLnkAOTd\nOrUZ/wK69Dzu4IvrN4vs9Nes8vbwPa/ddZEzGR0cQMt0JBkhk9kU/qwqUseP1QRJ\n5I1jR4g8aYPL/ke9K35PxZWuDp3U0UPAZ3PjFAh+5T+fc7gzCs9dPzSHloruU+gl\nFQIDAQAB\n-----END PUBLIC KEY-----"
    }
  ]
}`

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotNil(t, c.GetHandlers())
	})

	t.Run("test no coordination service error", func(t *testing.T) {
		c, err := New("domain", "origin", "", &mockprotocol.MockProvider{
			ServiceErr: fmt.Errorf("sample-error"),
		})
		require.Error(t, err)
		require.Nil(t, c)
		require.EqualError(t, err, "sample-error")
	})

	t.Run("test invalid coordination service error", func(t *testing.T) {
		c, err := New("domain", "origin", "", &mockprotocol.MockProvider{
			ServiceMap: map[string]interface{}{
				mediatorsvc.Coordination: "xyz",
			},
		})
		require.Error(t, err)
		require.Nil(t, c)
		require.EqualError(t, err, "cast service to route service failed")
	})
}

func TestCommand_CreateBlocDID(t *testing.T) {
	t.Run("test error from request", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBufferString("--"))
		require.Error(t, cmdErr)
		require.Equal(t, InvalidRequestErrorCode, cmdErr.Code())
		require.Equal(t, command.ValidationError, cmdErr.Type())
	})

	t.Run("bad didDoc", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		jwk := &jose.JWK{}
		jwk.Key = ed25519.PublicKey{}

		v, err := did.NewVerificationMethodFromJWK("id", "type", "c", jwk)
		require.NoError(t, err)

		jwk.Key = make(chan struct{})

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{
			DIDDocument: &did.Doc{
				VerificationMethod: []did.VerificationMethod{*v},
			},
		}}

		var b bytes.Buffer
		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBufferString("{}"))
		require.Empty(t, b.Bytes())
		require.NoError(t, cmdErr)
	})

	t.Run("test error unsupported purpose", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{}

		var b bytes.Buffer

		req, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{
			{
				ID: "key1", Type: "key1", KeyType: "Ed25519",
				Value:    base64.RawURLEncoding.EncodeToString([]byte("value")),
				Purposes: []string{"wrong"},
			},
		}})
		require.NoError(t, err)

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "public key purpose wrong not supported")
	})

	t.Run("test error from create did", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		req, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{
			{
				ID: "key1", Type: "key1", KeyType: "Ed25519",
				Value: base64.RawURLEncoding.EncodeToString([]byte("value")),
				Purposes: []string{
					doc.KeyPurposeAuthentication,
					doc.KeyPurposeKeyAgreement,
					doc.KeyPurposeCapabilityDelegation,
					doc.KeyPurposeCapabilityInvocation,
					doc.KeyPurposeAuthentication,
					doc.KeyPurposeAssertionMethod,
				},
			},
		}})
		require.NoError(t, err)

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "error create did")
	})

	t.Run("test recovery key not supported", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		req, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{
			{
				KeyType:  "wrong",
				Recovery: true,
			},
			{
				Type:  "key1",
				Value: "value",
			},
		}})
		require.NoError(t, err)

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "invalid key type: wrong")
	})

	t.Run("test error from did base64 decode", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		req, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{
			{
				Type:  "key1",
				Value: "value",
			},
		}})
		require.NoError(t, err)

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "illegal base64 data")
	})

	c, err := New("domain", "origin", "", getMockProvider())
	require.NoError(t, err)
	require.NotNil(t, c)

	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	ecPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	ecPubKeyBytes := elliptic.Marshal(ecPrivKey.PublicKey.Curve, ecPrivKey.PublicKey.X, ecPrivKey.PublicKey.Y)

	ec384PrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	ec384PubKeyBytes := elliptic.Marshal(ec384PrivKey.PublicKey.Curve, ec384PrivKey.PublicKey.X, ec384PrivKey.PublicKey.Y)

	bbsPubKey, _, err := bbs12381g2pub.GenerateKeyPair(sha256.New, nil)
	require.NoError(t, err)

	bbsPubKeyBytes, err := bbsPubKey.Marshal()
	require.NoError(t, err)

	t.Run("test success create did with Ed25519 key", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}
		// ED key
		r, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{
			{
				KeyType:  ed25519KeyType,
				Value:    base64.RawURLEncoding.EncodeToString(pubKey),
				Recovery: true,
			},
			{
				KeyType: p256KeyType,
				Value:   base64.RawURLEncoding.EncodeToString(ecPubKeyBytes),
				Update:  true,
			},
			{
				ID: "key1", Type: "key1", KeyType: "Ed25519",
				Value: base64.RawURLEncoding.EncodeToString([]byte("value")),
			},
		}})
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		docRes, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, docRes)
		require.Equal(t, "did:peer:21tDAKCERh95uGgKbJNHYp", docRes.DIDDocument.ID)
	})

	t.Run("test success create did with ecdsa p384 key", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}
		// ED key
		r, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{
			{
				KeyType:  ed25519KeyType,
				Value:    base64.RawURLEncoding.EncodeToString(pubKey),
				Recovery: true,
			},
			{
				KeyType: p256KeyType,
				Value:   base64.RawURLEncoding.EncodeToString(ecPubKeyBytes),
				Update:  true,
			},
			{
				KeyType: p384KeyType,
				Value:   base64.RawURLEncoding.EncodeToString(ec384PubKeyBytes),
			},
		}})
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		docRes, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, docRes)
		require.Equal(t, "did:peer:21tDAKCERh95uGgKbJNHYp", docRes.DIDDocument.ID)
	})

	t.Run("test success create did with BLS12381G2 key", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}
		// ED key
		r, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{
			{
				KeyType:  ed25519KeyType,
				Value:    base64.RawURLEncoding.EncodeToString(pubKey),
				Recovery: true,
			},
			{
				KeyType: p256KeyType,
				Value:   base64.RawURLEncoding.EncodeToString(ecPubKeyBytes),
				Update:  true,
			},
			{
				ID:      "key1",
				KeyType: BLS12381G2KeyType,
				Value:   base64.RawURLEncoding.EncodeToString(bbsPubKeyBytes),
			},
		}})
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		docRes, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, docRes)
		require.Equal(t, "did:peer:21tDAKCERh95uGgKbJNHYp", docRes.DIDDocument.ID)
	})
}

func TestCommand_CreatePeerDID(t *testing.T) {
	t.Run("test error from request", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString("--"))
		require.Error(t, cmdErr)
		require.Equal(t, InvalidRequestErrorCode, cmdErr.Code())
		require.Equal(t, command.ValidationError, cmdErr.Type())

		cmdErr = c.CreatePeerDID(&b, bytes.NewBufferString("{}"))
		require.Error(t, cmdErr)
		require.Equal(t, InvalidRequestErrorCode, cmdErr.Code())
		require.Equal(t, command.ValidationError, cmdErr.Type())
	})

	t.Run("success (registered route)", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		routerEndpoint := "http://router.com"
		keys := []string{"abc", "xyz"}
		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			ID:      uuid.NewString(),
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:              uuid.New().String(),
					Type:            didCommServiceType,
					ServiceEndpoint: routerEndpoint,
					RoutingKeys:     keys,
					RecipientKeys:   []string{"1ert5", "x5356s"},
				},
			},
		}}

		mediatorConfig := mediatorsvc.NewConfig(routerEndpoint, keys)
		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return mediatorConfig, nil
			},
		}

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Nil(t, cmdErr)

		resp, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, resp.DIDDocument)
		require.NotEmpty(t, resp.Context)
	})

	t.Run("success (default)", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			ID:      uuid.NewString(),
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:              uuid.New().String(),
					Type:            didCommServiceType,
					ServiceEndpoint: "http://router.com",
				},
			},
		}}

		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return &mediatorsvc.Config{}, nil
			},
		}

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Nil(t, cmdErr)

		resp, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, resp.DIDDocument)
		require.NotEmpty(t, resp.Context)
	})

	t.Run("test error while creating peer DID", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		// error while getting mediator config
		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return nil, fmt.Errorf("sample-error-1")
			},
		}

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Error(), "sample-error-1")

		// error while create peer DID from vdri
		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return &mediatorsvc.Config{}, nil
			},
		}

		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateErr: fmt.Errorf("sample-error-2")}

		cmdErr = c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Error(), "sample-error-2")

		// error for missing 'did-communication'
		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			Context: []string{"https://w3id.org/did/v1"},
		}}

		cmdErr = c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Error(), fmt.Sprintf(errMissingDIDCommServiceType, didCommServiceType))

		// error while adding router key
		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:              uuid.New().String(),
					Type:            didCommServiceType,
					ServiceEndpoint: "http://router.com",
					RecipientKeys:   []string{"1ert5", "x5356s"},
				},
			},
		}}

		c.mediatorSvc = &mockroute.MockMediatorSvc{
			AddKeyErr: fmt.Errorf("sample-error-3"),
		}

		cmdErr = c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))

		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "sample-error-3")
	})

	t.Run("test error while creating verification method", func(t *testing.T) {
		c, err := New("domain", "origin", "", getMockProvider())
		require.NoError(t, err)

		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:              uuid.New().String(),
					Type:            didCommServiceType,
					ServiceEndpoint: "http://router.com",
				},
			},
		}}

		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return &mediatorsvc.Config{}, nil
			},
		}

		c.keyManager = &mockkms.KeyManager{CrAndExportPubKeyErr: errors.New("test error")}

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.NotNil(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "test error")
	})
}

type mockDIDClient struct {
	createDIDValue *did.DocResolution
	createDIDErr   error
}

func (m *mockDIDClient) Create(didDoc *did.Doc, opts ...vdr.DIDMethodOption) (*did.DocResolution, error) {
	return m.createDIDValue, m.createDIDErr
}

// mockMediatorClient mock mediator client.
type mockMediatorClient struct {
	RegisterErr   error
	GetConfigFunc func(connID string) (*mediatorsvc.Config, error)
}

// Register registers with the router.
func (c *mockMediatorClient) Register(connectionID string) error {
	if c.RegisterErr != nil {
		return c.RegisterErr
	}

	return nil
}

// GetConfig gets the router config.
func (c *mockMediatorClient) GetConfig(connID string) (*mediatorsvc.Config, error) {
	return c.GetConfigFunc(connID)
}

func getMockProvider() Provider {
	return &mockprotocol.MockProvider{
		ServiceMap: map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{},
		},
	}
}
