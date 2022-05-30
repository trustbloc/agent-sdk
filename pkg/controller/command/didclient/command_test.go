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

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/sidetree/doc"
	"github.com/hyperledger/aries-framework-go/pkg/common/model"
	cryptoapi "github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/primitive/bbs12381g2pub"
	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
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
    },
	{
      "id":"did:peer:123456789abcdefghi#keys-3",
      "type":"X25519KeyAgreementKey2019",
      "controller": "did:peer:123456789abcdefghi",
      "publicKeyJwk": {"crv":"X25519","kty":"OKP","x":"yXh_D2YElByhNFDu-WkaE9NHcv0xcytantsJgAP07yA"}
	}
  ],
  "keyAgreement": [
  ]
}`

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotNil(t, c.GetHandlers())
	})

	t.Run("test no coordination service error", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, &mockprotocol.MockProvider{
			ServiceErr: fmt.Errorf("sample-error"),
		})
		require.Error(t, err)
		require.Nil(t, c)
		require.EqualError(t, err, "sample-error")
	})

	t.Run("test invalid coordination service error", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, &mockprotocol.MockProvider{
			ServiceMap: map[string]interface{}{
				mediatorsvc.Coordination: "xyz",
			},
		})
		require.Error(t, err)
		require.Nil(t, c)
		require.EqualError(t, err, "cast service to route service failed")
	})
}

func TestCommand_ResolveOrbDID(t *testing.T) {
	t.Run("test error from request", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer

		cmdErr := c.ResolveOrbDID(&b, bytes.NewBufferString("--"))
		require.Error(t, cmdErr)
		require.Equal(t, InvalidRequestErrorCode, cmdErr.Code())
		require.Equal(t, command.ValidationError, cmdErr.Type())
	})

	t.Run("test error from resolve did", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{resolveDIDErr: fmt.Errorf("error resolve did")}

		var b bytes.Buffer

		req, err := json.Marshal(ResolveOrbDIDRequest{DID: "did:123"})
		require.NoError(t, err)

		cmdErr := c.ResolveOrbDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, ResolveDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "error resolve did")
	})

	t.Run("test success", func(t *testing.T) {
		c, err := New("domain", "origin", "", 1, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{resolveDIDValue: &did.DocResolution{DIDDocument: &did.Doc{
			ID:      "did:123",
			Context: []string{"https://www.w3.org/ns/did/v1"},
		}}}

		req, err := json.Marshal(ResolveOrbDIDRequest{DID: "did:123"})
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.ResolveOrbDID(&b, bytes.NewBuffer(req))
		require.NoError(t, cmdErr)

		docRes, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, docRes)
		require.Contains(t, "did:123", docRes.DIDDocument.ID)
	})
}

func TestCommand_CreateOrbDID(t *testing.T) {
	t.Run("test error from request", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		cmdErr := c.CreateOrbDID(&b, bytes.NewBufferString("--"))
		require.Error(t, cmdErr)
		require.Equal(t, InvalidRequestErrorCode, cmdErr.Code())
		require.Equal(t, command.ValidationError, cmdErr.Type())
	})

	t.Run("bad didDoc", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		j := &jwk.JWK{}
		j.Key = ed25519.PublicKey{}

		v, err := did.NewVerificationMethodFromJWK("id", "type", "c", j)
		require.NoError(t, err)

		j.Key = make(chan struct{})

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{
			DIDDocument: &did.Doc{
				VerificationMethod: []did.VerificationMethod{*v},
			},
		}}

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBufferString("{}"))
		require.Empty(t, b.Bytes())
		require.Error(t, cmdErr)
	})

	t.Run("test error unsupported purpose", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{}

		var b bytes.Buffer

		req, err := json.Marshal(CreateOrbDIDRequest{PublicKeys: []PublicKey{
			{
				ID: "key1", Type: "key1", KeyType: "Ed25519",
				Value:    base64.RawURLEncoding.EncodeToString([]byte("value")),
				Purposes: []string{"wrong"},
			},
		}})
		require.NoError(t, err)

		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "public key purpose wrong not supported")
	})

	t.Run("test error from create did", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		req, err := json.Marshal(CreateOrbDIDRequest{PublicKeys: []PublicKey{
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

		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "error create did")
	})

	t.Run("test recovery key not supported", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		req, err := json.Marshal(CreateOrbDIDRequest{PublicKeys: []PublicKey{
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

		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "invalid key type: wrong")
	})

	t.Run("test error from did base64 decode", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		req, err := json.Marshal(CreateOrbDIDRequest{PublicKeys: []PublicKey{
			{
				Type:  "key1",
				Value: "value",
			},
		}})
		require.NoError(t, err)

		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "illegal base64 data")
	})

	c, err := New("domain", "origin", "", 0, getMockProvider())
	require.NoError(t, err)
	require.NotNil(t, c)

	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	ecPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	ecPubKeyBytes := elliptic.Marshal(ecPrivKey.PublicKey.Curve, ecPrivKey.PublicKey.X, ecPrivKey.PublicKey.Y)

	nistP256ECDHKWKeyBytes, err := json.Marshal(&cryptoapi.PublicKey{
		X:     ecPrivKey.PublicKey.X.Bytes(),
		Y:     ecPrivKey.PublicKey.Y.Bytes(),
		Curve: ecPrivKey.PublicKey.Curve.Params().Name,
		Type:  "ec",
	})
	require.NoError(t, err)

	ec384PrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	ec384PubKeyBytes := elliptic.Marshal(ec384PrivKey.PublicKey.Curve, ec384PrivKey.PublicKey.X, ec384PrivKey.PublicKey.Y)

	nistP384ECDHKWKeyBytes, err := json.Marshal(&cryptoapi.PublicKey{
		X:     ec384PrivKey.PublicKey.X.Bytes(),
		Y:     ec384PrivKey.PublicKey.Y.Bytes(),
		Curve: ec384PrivKey.PublicKey.Curve.Params().Name,
		Type:  "ec",
	})
	require.NoError(t, err)

	ec521PrivKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	require.NoError(t, err)

	nistP521ECDHKWKeyBytes, err := json.Marshal(&cryptoapi.PublicKey{
		X:     ec521PrivKey.PublicKey.X.Bytes(),
		Y:     ec521PrivKey.PublicKey.Y.Bytes(),
		Curve: ec521PrivKey.PublicKey.Curve.Params().Name,
		Type:  "ec",
	})
	require.NoError(t, err)

	badNISTP384ECDHKWKeyBytes, err := json.Marshal(&cryptoapi.PublicKey{
		X:     []byte{'`'},
		Y:     []byte{'`'},
		Curve: ec384PrivKey.PublicKey.Curve.Params().Name,
		Type:  "badType",
	})
	require.NoError(t, err)

	bbsPubKey, _, err := bbs12381g2pub.GenerateKeyPair(sha256.New, nil)
	require.NoError(t, err)

	bbsPubKeyBytes, err := bbsPubKey.Marshal()
	require.NoError(t, err)

	x25519PublicKey, err := json.Marshal(&cryptoapi.PublicKey{
		X:     pubKey,
		Curve: "X25519",
		Type:  "okp",
	})
	require.NoError(t, err)

	badX25519PublicKey, err := json.Marshal(&cryptoapi.PublicKey{
		X:     []byte{'`'},
		Curve: "X25519",
		Type:  "badType",
	})
	require.NoError(t, err)

	t.Run("test success create did with Ed25519 key as main and X25519 key as KeyAgreement", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}
		// ED key
		r, err := json.Marshal(CreateOrbDIDRequest{
			PublicKeys: []PublicKey{
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
				{
					KeyType:  x25519ECDHKW,
					Value:    base64.RawURLEncoding.EncodeToString(x25519PublicKey),
					Purposes: []string{"keyAgreement"},
				},
			},
		})
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
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
		r, err := json.Marshal(CreateOrbDIDRequest{PublicKeys: []PublicKey{
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
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
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
		r, err := json.Marshal(CreateOrbDIDRequest{PublicKeys: []PublicKey{
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
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		docRes, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, docRes)
		require.Equal(t, "did:peer:21tDAKCERh95uGgKbJNHYp", docRes.DIDDocument.ID)
	})

	t.Run("test success create did with custom properties, P384ECDHKW keyAgreement + router service", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)
		didDoc.KeyAgreement = []did.Verification{{
			VerificationMethod: didDoc.VerificationMethod[2],
			Relationship:       did.KeyAgreement,
		}}
		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}

		r, err := json.Marshal(
			CreateOrbDIDRequest{
				PublicKeys: []PublicKey{
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
						KeyType:  p384ecdhkw,
						Value:    base64.RawURLEncoding.EncodeToString(nistP384ECDHKWKeyBytes),
						Purposes: []string{"keyAgreement"},
					},
				},
				DIDcommServiceType: "did-communication",
				ServiceID:          "testService",
				ServiceEndpoint:    "http//test.serviceEndpoint",
				RouterConnections:  []string{"12345"},
				RoutersKeyAgrIDS:   []string{"12345"},
			},
		)
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		docRes, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, docRes)
		require.Equal(t, "did:peer:21tDAKCERh95uGgKbJNHYp", docRes.DIDDocument.ID)
	})

	t.Run("test fail create did with custom properties, using bad router service", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)
		didDoc.KeyAgreement = []did.Verification{{
			VerificationMethod: didDoc.VerificationMethod[2],
			Relationship:       did.KeyAgreement,
		}}

		addRouterKeyErr := fmt.Errorf("add router key failed")

		badC, err := New("domain", "origin", "", 0,
			getMockProviderWithMediator(&mockroute.MockMediatorSvc{
				AddKeyErr: addRouterKeyErr,
			}))
		require.NoError(t, err)
		require.NotNil(t, c)

		badC.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}

		r, err := json.Marshal(
			CreateOrbDIDRequest{
				PublicKeys: []PublicKey{
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
						KeyType:  p384ecdhkw,
						Value:    base64.RawURLEncoding.EncodeToString(nistP384ECDHKWKeyBytes),
						Purposes: []string{"keyAgreement"},
					},
				},
				DIDcommServiceType: "did-communication",
				ServiceID:          "testService",
				ServiceEndpoint:    "http//test.serviceEndpoint",
				RouterConnections:  []string{"12345"},
				RoutersKeyAgrIDS:   []string{"12345"},
			},
		)
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := badC.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.Contains(t, cmdErr.Error(), "failed to register did doc recipient key")
		require.True(t, errors.As(cmdErr, &addRouterKeyErr))
	})

	t.Run("test success create did with custom properties and NISTP256ECDHKW keyAgreement", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}

		r, err := json.Marshal(
			CreateOrbDIDRequest{
				PublicKeys: []PublicKey{
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
						KeyType:  p256ecdhkw,
						Value:    base64.RawURLEncoding.EncodeToString(nistP256ECDHKWKeyBytes),
						Purposes: []string{"keyAgreement"},
					},
				},
				DIDcommServiceType: "did-communication",
				ServiceID:          "testService",
				ServiceEndpoint:    "http//test.serviceEndpoint",
				RouterConnections:  []string{"12345"},
				RoutersKeyAgrIDS:   []string{"12345"},
			},
		)
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		docRes, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, docRes)
		require.Equal(t, "did:peer:21tDAKCERh95uGgKbJNHYp", docRes.DIDDocument.ID)
	})

	t.Run("test success create did with custom properties and NISTP521ECDHKW keyAgreement", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}

		r, err := json.Marshal(
			CreateOrbDIDRequest{
				PublicKeys: []PublicKey{
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
						KeyType:  p521ecdhkw,
						Value:    base64.RawURLEncoding.EncodeToString(nistP521ECDHKWKeyBytes),
						Purposes: []string{"keyAgreement"},
					},
				},
				DIDcommServiceType: "did-communication",
				ServiceID:          "testService",
				ServiceEndpoint:    "http//test.serviceEndpoint",
				RouterConnections:  []string{"12345"},
				RoutersKeyAgrIDS:   []string{"12345"},
			},
		)
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		docRes, err := did.ParseDocumentResolution(b.Bytes())
		require.NoError(t, err)
		require.NotEmpty(t, docRes)
		require.Equal(t, "did:peer:21tDAKCERh95uGgKbJNHYp", docRes.DIDDocument.ID)
	})

	t.Run("test fail create did with custom properties and bad x25519 key", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}

		r, err := json.Marshal(
			CreateOrbDIDRequest{
				PublicKeys: []PublicKey{
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
						KeyType:  x25519ECDHKW,
						Value:    base64.RawURLEncoding.EncodeToString(badX25519PublicKey),
						Purposes: []string{"keyAgreement"},
					},
				},
			},
		)
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.EqualError(t, cmdErr, "create JWK: marshalX25519: invalid key")
	})

	t.Run("test fail to create did with custom properties with invalid P-384 ecdsa key type", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}

		r, err := json.Marshal(
			CreateOrbDIDRequest{
				PublicKeys: []PublicKey{
					{
						KeyType:  ed25519KeyType,
						Value:    base64.RawURLEncoding.EncodeToString(pubKey),
						Recovery: true,
					},
					{
						KeyType: p384KeyType,
						Value:   base64.RawURLEncoding.EncodeToString(ecPubKeyBytes), // ecPubKeyBytes is p-256
					},
					{
						KeyType:  x25519ECDHKW,
						Value:    base64.RawURLEncoding.EncodeToString(x25519PublicKey),
						Purposes: []string{"keyAgreement"},
					},
				},
				DIDcommServiceType: "did-communication",
				ServiceID:          "testService",
				ServiceEndpoint:    "http//test.serviceEndpoint",
			},
		)
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.EqualError(t, cmdErr, "create JWK: square/go-jose: invalid EC key (nil, or X/Y missing)")
	})

	t.Run("test fail create did with custom properties and bad NISTP384ECDHKW public key", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}

		r, err := json.Marshal(
			CreateOrbDIDRequest{
				PublicKeys: []PublicKey{
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
						KeyType:  p384ecdhkw,
						Value:    base64.RawURLEncoding.EncodeToString(badNISTP384ECDHKWKeyBytes),
						Purposes: []string{"keyAgreement"},
					},
				},
				DIDcommServiceType: "did-communication",
				ServiceID:          "testService",
				ServiceEndpoint:    "http//test.serviceEndpoint",
			},
		)
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.Contains(t, cmdErr.Error(), "JWKFromKey() jwk: <nil>, ecdsa key:")
	})

	t.Run("test fail create did with custom properties and NISTP384ECDHKW as not public key", func(t *testing.T) {
		didDoc, err := did.ParseDocument([]byte(sampleDoc))
		require.NoError(t, err)

		c.didBlocClient = &mockDIDClient{createDIDValue: &did.DocResolution{DIDDocument: didDoc}}

		r, err := json.Marshal(
			CreateOrbDIDRequest{
				PublicKeys: []PublicKey{
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
						KeyType:  p384ecdhkw,
						Value:    base64.RawURLEncoding.EncodeToString([]byte("bad key, not *cryptoapi.PublicKey{}")),
						Purposes: []string{"keyAgreement"},
					},
				},
				DIDcommServiceType: "did-communication",
				ServiceID:          "testService",
				ServiceEndpoint:    "http//test.serviceEndpoint",
			},
		)
		require.NoError(t, err)

		var b bytes.Buffer
		cmdErr := c.CreateOrbDID(&b, bytes.NewBuffer(r))
		require.Contains(t, cmdErr.Error(), "unmarshal key type: nistp384ecdhkw, value: bad key, not "+
			"*cryptoapi.PublicKey{} failed: invalid character")
	})
}

func TestCommand_CreatePeerDID(t *testing.T) {
	t.Run("test error from request", func(t *testing.T) {
		c, err := New("domain", "origin", "", 0, getMockProvider())
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
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		routerEndpoint := "http://router.com"
		keys := []string{"abc", "xyz"}
		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			ID:      uuid.NewString(),
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:   uuid.New().String(),
					Type: didCommServiceType,
					ServiceEndpoint: model.NewDIDCommV2Endpoint(
						[]model.DIDCommV2Endpoint{{URI: routerEndpoint}}),
					RoutingKeys:   keys,
					RecipientKeys: []string{"1ert5", "x5356s"},
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
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			ID:      uuid.NewString(),
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:   uuid.New().String(),
					Type: didCommServiceType,
					ServiceEndpoint: model.NewDIDCommV2Endpoint(
						[]model.DIDCommV2Endpoint{{URI: "http://router.com"}}),
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
		c, err := New("domain", "origin", "", 0, getMockProvider())
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
					ID:   uuid.New().String(),
					Type: didCommServiceType,
					ServiceEndpoint: model.NewDIDCommV2Endpoint(
						[]model.DIDCommV2Endpoint{{URI: "http://router.com"}}),
					RecipientKeys: []string{"1ert5", "x5356s"},
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
		c, err := New("domain", "origin", "", 0, getMockProvider())
		require.NoError(t, err)

		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:   uuid.New().String(),
					Type: didCommServiceType,
					ServiceEndpoint: model.NewDIDCommV2Endpoint(
						[]model.DIDCommV2Endpoint{{URI: "http://router.com"}}),
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
	createDIDValue  *did.DocResolution
	createDIDErr    error
	resolveDIDValue *did.DocResolution
	resolveDIDErr   error
}

func (m *mockDIDClient) Create(didDoc *did.Doc, opts ...vdr.DIDMethodOption) (*did.DocResolution, error) {
	return m.createDIDValue, m.createDIDErr
}

func (m *mockDIDClient) Read(id string, opts ...vdr.DIDMethodOption) (*did.DocResolution, error) {
	return m.resolveDIDValue, m.resolveDIDErr
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
	return getMockProviderWithMediator(&mockroute.MockMediatorSvc{})
}

func getMockProviderWithMediator(mediator interface{}) Provider {
	return &mockprotocol.MockProvider{
		ServiceMap: map[string]interface{}{
			mediatorsvc.Coordination: mediator,
		},
	}
}
