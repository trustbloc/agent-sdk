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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/sidetree/doc"
	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	mockprotocol "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockvdr "github.com/hyperledger/aries-framework-go/pkg/mock/vdr"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
)

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New("domain", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotNil(t, c.GetHandlers())
	})

	t.Run("test no coordination service error", func(t *testing.T) {
		c, err := New("domain", &mockprotocol.MockProvider{
			ServiceErr: fmt.Errorf("sample-error"),
		})
		require.Error(t, err)
		require.Nil(t, c)
		require.EqualError(t, err, "sample-error")
	})

	t.Run("test invalid coordination service error", func(t *testing.T) {
		c, err := New("domain", &mockprotocol.MockProvider{
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
		c, err := New("domain", getMockProvider())
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
		c, err := New("domain", getMockProvider())
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
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
	})

	t.Run("test error unsupported purpose", func(t *testing.T) {
		c, err := New("domain", getMockProvider())
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
		c, err := New("domain", getMockProvider())
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
		c, err := New("domain", getMockProvider())
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
		c, err := New("domain", getMockProvider())
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

	c, err := New("domain", getMockProvider())
	require.NoError(t, err)
	require.NotNil(t, c)

	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	ecPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	ecPubKeyBytes := elliptic.Marshal(ecPrivKey.PublicKey.Curve, ecPrivKey.PublicKey.X, ecPrivKey.PublicKey.Y)

	c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: &did.Doc{ID: "1"}}}

	var b bytes.Buffer

	t.Run("test success create did with Ed25519 key", func(t *testing.T) {
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

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		resp := &CreateDIDResponse{}
		err = json.NewDecoder(&b).Decode(&resp)
		require.NoError(t, err)

		var didMap map[string]string
		err = json.Unmarshal(resp.DID, &didMap)
		require.NoError(t, err)

		require.Equal(t, "1", didMap["id"])
	})
}

func TestCommand_CreatePeerDID(t *testing.T) {
	t.Run("test error from request", func(t *testing.T) {
		c, err := New("domain", getMockProvider())
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
		c, err := New("domain", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		routerEndpoint := "http://router.com"
		keys := []string{"abc", "xyz"}
		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
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

		resp := &CreateDIDResponse{}
		err = json.NewDecoder(&b).Decode(&resp)
		require.NoError(t, err)

		var didMap map[string]interface{}
		err = json.Unmarshal(resp.DID, &didMap)
		require.NoError(t, err)
		require.NotEmpty(t, didMap["service"])
	})

	t.Run("success (default)", func(t *testing.T) {
		c, err := New("domain", getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

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

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Nil(t, cmdErr)

		resp := &CreateDIDResponse{}
		err = json.NewDecoder(&b).Decode(&resp)
		require.NoError(t, err)

		var didMap map[string]interface{}
		err = json.Unmarshal(resp.DID, &didMap)
		require.NoError(t, err)
		require.NotEmpty(t, didMap["service"])
	})

	t.Run("test error while creating peer DID", func(t *testing.T) {
		c, err := New("domain", getMockProvider())
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
}

type mockDIDClient struct {
	createDIDValue *did.DocResolution
	createDIDErr   error
}

func (m *mockDIDClient) Create(keyManager kms.KeyManager, didDoc *did.Doc,
	opts ...vdr.DIDMethodOption) (*did.DocResolution, error) {
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
