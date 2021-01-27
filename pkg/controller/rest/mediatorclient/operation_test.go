/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mediatorclient // nolint:testpackage // uses internal implementation details

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexchangesvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
	mockmsghandler "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/msghandler"
	mockdidexchange "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockkms "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	mockstorage "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/mediatorclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks"
	sdkmockprotocol "github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks/protocol"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/testutil"
)

const sampleErr = "sample-error"

//nolint:lll
const sampleDIDDoc = `{
  "@context": ["https://w3id.org/did/v1"],
  "id": "did:example:21tDAKCERh95uGgKbJNHYp",
  "verificationMethod": [
    {
      "id": "did:example:123456789abcdefghi#keys-1",
      "type": "Secp256k1VerificationKey2018",
      "controller": "did:example:123456789abcdefghi",
      "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
    },
    {
      "id": "did:example:123456789abcdefghw#key2",
      "type": "RsaVerificationKey2018",
      "controller": "did:example:123456789abcdefghw",
      "publicKeyPem": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAryQICCl6NZ5gDKrnSztO\n3Hy8PEUcuyvg/ikC+VcIo2SFFSf18a3IMYldIugqqqZCs4/4uVW3sbdLs/6PfgdX\n7O9D22ZiFWHPYA2k2N744MNiCD1UE+tJyllUhSblK48bn+v1oZHCM0nYQ2NqUkvS\nj+hwUU3RiWl7x3D2s9wSdNt7XUtW05a/FXehsPSiJfKvHJJnGOX0BgTvkLnkAOTd\nOrUZ/wK69Dzu4IvrN4vs9Nes8vbwPa/ddZEzGR0cQMt0JBkhk9kU/qwqUseP1QRJ\n5I1jR4g8aYPL/ke9K35PxZWuDp3U0UPAZ3PjFAh+5T+fc7gzCs9dPzSHloruU+gl\nFQIDAQAB\n-----END PUBLIC KEY-----"
    }
  ],
  "authentication": [
    "did:example:123456789abcdefghi#keys-1",
    {
      "id": "did:example:123456789abcdefghs#key3",
      "type": "RsaVerificationKey2018",
      "controller": "did:example:123456789abcdefghs",
      "publicKeyHex": "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71"
    }
  ],
  "service": [
    {
      "id": "did:example:123456789abcdefghi#inbox",
      "type": "SocialWebInboxService",
      "serviceEndpoint": "https://social.example.com/83hfh37dj",
      "spamCost": {
        "amount": "0.50",
        "currency": "USD"
      }
    },
    {
      "id": "did:example:123456789abcdefghi#did-communication",
      "type": "did-communication",
      "serviceEndpoint": "https://agent.example.com/",
      "priority" : 0,
      "recipientKeys" : ["did:example:123456789abcdefghi#key2"],
      "routingKeys" : ["did:example:123456789abcdefghi#key2"]
    }
  ],
  "created": "2002-10-10T17:00:00Z"
}`

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New(newMockProvider(nil), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotEmpty(t, c.GetRESTHandlers())
		require.Len(t, c.GetRESTHandlers(), 3)
	})

	t.Run("test failure while creating mediator client", func(t *testing.T) {
		c, err := New(sdkmockprotocol.NewMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create mediator client")
	})
}

func TestOperation_Connect(t *testing.T) {
	const (
		sampleConnID         = "sample-conn-id"
		sampleRouterEndpoint = "sample-router-endpoint"
		sampleRoutingKeys    = "sample-routing-keys"
		sampleInvitation     = `{
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
	)

	t.Run("test successful connect", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				RouterEndpoint: sampleRouterEndpoint,
				RoutingKeys:    []string{sampleRoutingKeys},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{ConnID: sampleConnID},
			outofbandsvc.Name: &sdkmockprotocol.MockOobService{
				AcceptInvitationHandle: func(_ *outofbandsvc.Invitation, _ string, _ []string) (s string, e error) {
					return sampleConnID, nil
				},
			},
		})

		cmd, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, ConnectPath)

		buf, err := testutil.GetSuccessResponseFromHandler(handler, bytes.NewBufferString(sampleInvitation), handler.Path())
		require.NoError(t, err)

		resp := &connectionResponse{}
		err = json.NewDecoder(buf).Decode(&resp.Response)
		require.NoError(t, err)

		require.Equal(t, resp.Response.ConnectionID, sampleConnID)
	})

	t.Run("test failed connect", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination:   &mockroute.MockMediatorSvc{},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name: &sdkmockprotocol.MockOobService{
				AcceptInvitationHandle: func(_ *outofbandsvc.Invitation, _ string, _ []string) (s string, e error) {
					return "", fmt.Errorf(sampleErr)
				},
			},
		})

		cmd, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, ConnectPath)

		buf, code, err := testutil.SendRequestToHandler(handler, bytes.NewBufferString(sampleInvitation), handler.Path())
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		require.Equal(t, http.StatusInternalServerError, code)
		testutil.VerifyError(t, mediatorclient.ConnectMediatorError, sampleErr, buf.Bytes())
	})
}

func TestOperation_CreateInvitation(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				Connections: []string{"sample-connection"},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		cmd, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, CreateInvitationPath)

		_, err = testutil.GetSuccessResponseFromHandler(handler, bytes.NewBufferString("{}"), handler.Path())
		require.NoError(t, err)
	})

	t.Run("test failure", func(t *testing.T) {
		cmd, err := New(newMockProvider(nil), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, CreateInvitationPath)

		buf, code, err := testutil.SendRequestToHandler(handler, bytes.NewBufferString(""), handler.Path())
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		require.Equal(t, http.StatusInternalServerError, code)
		testutil.VerifyError(t, mediatorclient.CreateInvitationError,
			"no connection found to create invitation", buf.Bytes())
	})
}

func TestOperation_SendCreateConnectionRequest(t *testing.T) {
	const replyMsgStr = `{
							"@id": "123456781",
							"@type": "https://trustbloc.dev/blinded-routing/1.0/create-conn-resp",
							"~thread" : {"thid": "%s"}
					}`

	t.Run("test success", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				Connections: []string{"sample-connection"},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		record := &connection.Record{
			ConnectionID: "sample-connection",
			State:        "completed", MyDID: "mydid", TheirDID: "theirDID-001",
		}
		mockStore := &mockstorage.MockStore{Store: make(map[string][]byte)}
		connBytes, err := json.Marshal(record)
		require.NoError(t, err)
		require.NoError(t, mockStore.Put("conn_sample-connection", connBytes))
		prov.StoreProvider = mockstorage.NewCustomMockStoreProvider(mockStore)

		registrar := mockmsghandler.NewMockMsgServiceProvider()
		mockmsgr := sdkmockprotocol.NewMockMessenger()

		prov.CustomMessenger = mockmsgr

		go func() {
			for {
				if len(registrar.Services()) > 0 && mockmsgr.GetLastID() != "" { //nolint: gocritic
					replyMsg, e := service.ParseDIDCommMsgMap([]byte(fmt.Sprintf(replyMsgStr, mockmsgr.GetLastID())))
					require.NoError(t, e)

					_, e = registrar.Services()[0].HandleInbound(replyMsg, "sampleDID", "sampleTheirDID")
					require.NoError(t, e)

					break
				}
			}
		}()

		cmd, err := New(prov, registrar, mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, SendCreateConnectionRequest)

		request := createConnectionRequest{
			Request: mediatorclient.CreateConnectionRequest{
				DIDDocument: json.RawMessage([]byte(sampleDIDDoc)),
			},
		}

		rqstBytes, err := json.Marshal(request)
		require.NoError(t, err)

		_, err = testutil.GetSuccessResponseFromHandler(handler, bytes.NewBuffer(rqstBytes), handler.Path())
		require.NoError(t, err)
	})

	t.Run("test failure", func(t *testing.T) {
		cmd, err := New(newMockProvider(nil), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, SendCreateConnectionRequest)

		buf, code, err := testutil.SendRequestToHandler(handler, bytes.NewBufferString("---"), handler.Path())
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		require.Equal(t, http.StatusInternalServerError, code)
		testutil.VerifyError(t, mediatorclient.SendCreateConnectionRequestError,
			"no connection found to create invitation", buf.Bytes())
	})
}

func newMockProvider(serviceMap map[string]interface{}) *sdkmockprotocol.MockProvider {
	if serviceMap == nil {
		serviceMap = map[string]interface{}{
			mediatorsvc.Coordination:   &mockroute.MockMediatorSvc{},
			didexchangesvc.DIDExchange: &mockdidexchange.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		}
	}

	prov := sdkmockprotocol.NewMockProvider()

	prov.ServiceMap = serviceMap
	prov.StoreProvider = mockstorage.NewMockStoreProvider()
	prov.ProtocolStateStoreProvider = mockstorage.NewMockStoreProvider()
	prov.CustomKMS = &mockkms.KeyManager{}

	return prov
}
