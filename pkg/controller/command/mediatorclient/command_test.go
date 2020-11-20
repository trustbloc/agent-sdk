/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package mediatorclient // nolint:testpackage // uses internal implementation details

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexchangesvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
	mockmsghandler "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/msghandler"
	mockdidexchange "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockservice "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/service"
	mockkms "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	mockstorage "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks"
	sdkmockprotocol "github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks/protocol"
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
		require.NotEmpty(t, c.GetHandlers())
		require.Len(t, c.GetHandlers(), 3)
	})

	t.Run("test failure while creating mediator client", func(t *testing.T) {
		c, err := New(sdkmockprotocol.NewMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create mediator client")
	})

	t.Run("test failure while creating did-exchange client", func(t *testing.T) {
		c, err := New(newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{},
		}), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create did-exchange client")
	})

	t.Run("test failure while creating out-of-band client", func(t *testing.T) {
		c, err := New(newMockProvider(
			map[string]interface{}{
				mediatorsvc.Coordination:   &mockroute.MockMediatorSvc{},
				didexchangesvc.DIDExchange: &mockdidexchange.MockDIDExchangeSvc{},
			}), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create out-of-band client")
	})

	t.Run("test failure while creating messaging client", func(t *testing.T) {
		prov := newMockProvider(nil)
		prov.StoreProvider = &mockstorage.MockStoreProvider{
			FailNamespace: "didexchange",
		}

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create messenger client")
	})
}

func TestCommand_Connect(t *testing.T) {
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
		sampleInvitation2 = `{
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
		"mylabel": "sample-agent-label",
		"stateCompleteMessageType": "https://trustbloc.dev/didexchange/1.0/state-complete"
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation))
		require.NoError(t, cmdErr)

		resp := &ConnectionResponse{}
		err = json.NewDecoder(&b).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, resp.ConnectionID, sampleConnID)
		require.Equal(t, resp.RoutingKeys, []string{sampleRoutingKeys})
		require.Equal(t, resp.RouterEndpoint, sampleRouterEndpoint)
	})

	t.Run("test successful connect with state complete notification", func(t *testing.T) {
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

		mockMsgRegistrar := mockmsghandler.NewMockMsgServiceProvider()

		go func() {
			for {
				if len(mockMsgRegistrar.Services()) > 0 {
					_, e := mockMsgRegistrar.Services()[0].HandleInbound(&service.DIDCommMsgMap{}, "", "")
					require.NoError(t, e)

					break
				}
			}
		}()

		c, err := New(prov, mockMsgRegistrar, mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		// reduce timeout
		c.didExchTimeout = 10 * time.Millisecond

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation2))
		require.NoError(t, cmdErr)

		resp := &ConnectionResponse{}
		err = json.NewDecoder(&b).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, resp.ConnectionID, sampleConnID)
		require.Equal(t, resp.RoutingKeys, []string{sampleRoutingKeys})
		require.Equal(t, resp.RouterEndpoint, sampleRouterEndpoint)
	})

	t.Run("test failure due to incorrect request", func(t *testing.T) {
		prov := newMockProvider(nil)

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(`--`))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "invalid character")
		require.Equal(t, cmdErr.Type(), command.ValidationError)
		require.Equal(t, cmdErr.Code(), InvalidRequestErrorCode)
	})

	t.Run("test failure due to missing invitation", func(t *testing.T) {
		prov := newMockProvider(nil)

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(`{"mylabel":"test"}`))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "invitation missing in connection request")
		require.Equal(t, cmdErr.Type(), command.ValidationError)
		require.Equal(t, cmdErr.Code(), InvalidRequestErrorCode)
	})

	t.Run("test failure due to failed message event registration", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{},
			didexchangesvc.DIDExchange: &mockdidexchange.MockDIDExchangeSvc{
				RegisterMsgEventErr: fmt.Errorf(sampleErr),
			},
			outofbandsvc.Name: &sdkmockprotocol.MockOobService{},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), sampleErr)
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Equal(t, cmdErr.Code(), ConnectMediatorError)
	})

	t.Run("test failure due to accept invitation failure", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination:   &mockroute.MockMediatorSvc{},
			didexchangesvc.DIDExchange: &mockdidexchange.MockDIDExchangeSvc{},
			outofbandsvc.Name: &sdkmockprotocol.MockOobService{
				AcceptInvitationHandle: func(_ *outofbandsvc.Invitation, _ string, _ []string) (s string, e error) {
					return "", fmt.Errorf(sampleErr)
				},
			},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), sampleErr)
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Equal(t, cmdErr.Code(), ConnectMediatorError)
	})

	t.Run("test failure due to mediator registration error", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				RegisterFunc: func(connectionID string, _ ...mediatorsvc.ClientOption) error {
					return fmt.Errorf(sampleErr)
				},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{ConnID: sampleConnID},
			outofbandsvc.Name: &sdkmockprotocol.MockOobService{
				AcceptInvitationHandle: func(_ *outofbandsvc.Invitation, _ string, _ []string) (s string, e error) {
					return sampleConnID, nil
				},
			},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), sampleErr)
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Equal(t, cmdErr.Code(), ConnectMediatorError)
	})

	t.Run("test failure due to mediator registration error", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination:   &mockroute.MockMediatorSvc{},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{ConnID: sampleConnID},
			outofbandsvc.Name: &sdkmockprotocol.MockOobService{
				AcceptInvitationHandle: func(_ *outofbandsvc.Invitation, _ string, _ []string) (s string, e error) {
					return sampleConnID, nil
				},
			},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "router not registered")
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Equal(t, cmdErr.Code(), ConnectMediatorError)
	})

	t.Run("test failure due to didexchange state issues", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				RouterEndpoint: sampleRouterEndpoint,
				RoutingKeys:    []string{sampleRoutingKeys},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{
				ConnID: sampleConnID,
				State:  didexchangesvc.StateIDRequested,
			},
			outofbandsvc.Name: &sdkmockprotocol.MockOobService{
				AcceptInvitationHandle: func(_ *outofbandsvc.Invitation, _ string, _ []string) (s string, e error) {
					return sampleConnID, nil
				},
			},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		// reduce timeout
		c.didExchTimeout = 10 * time.Millisecond

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "time out waiting for did exchange state 'completed'")
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Equal(t, cmdErr.Code(), ConnectMediatorError)
	})

	t.Run("test failure messenger registration error", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination:   &mockroute.MockMediatorSvc{},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		c, err := New(prov, &mockmsghandler.MockMsgSvcProvider{
			RegisterErr: fmt.Errorf(sampleErr),
		}, mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		// reduce timeout
		c.didExchTimeout = 10 * time.Millisecond

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation2))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), sampleErr)
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Equal(t, cmdErr.Code(), ConnectMediatorError)
	})

	t.Run("test failure waiting for state complete notification", func(t *testing.T) {
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

		c, err := New(prov, &mockmsghandler.MockMsgSvcProvider{UnregisterErr: fmt.Errorf(sampleErr)}, mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		// reduce timeout
		c.didExchTimeout = 10 * time.Millisecond

		var b bytes.Buffer
		cmdErr := c.Connect(&b, bytes.NewBufferString(sampleInvitation2))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "timeout waiting for state completed message from mediator")
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Equal(t, cmdErr.Code(), ConnectMediatorError)
	})
}

func TestCommand_CreateInvitation(t *testing.T) {
	t.Run("test with empty connections", func(t *testing.T) {
		c, err := New(newMockProvider(nil), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.CreateInvitation(&b, bytes.NewBufferString(""))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "no connection found to create invitation")
	})

	t.Run("test failure while getting connections", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				GetConnectionsErr: fmt.Errorf(sampleErr),
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.CreateInvitation(&b, bytes.NewBufferString(""))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), sampleErr)
	})

	t.Run("test failure due to invalid request", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				Connections: []string{"sample-connection"},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.CreateInvitation(&b, bytes.NewBufferString("="))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "invalid character")
	})

	t.Run("test success", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				Connections: []string{"sample-connection"},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.CreateInvitation(&b, bytes.NewBufferString("{}"))
		require.NoError(t, cmdErr)
	})

	t.Run("test failure while saving invitation", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				Connections: []string{"sample-connection"},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name: &sdkmockprotocol.MockOobService{
				SaveInvitationErr: fmt.Errorf(sampleErr),
			},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		cmdErr := c.CreateInvitation(&b, bytes.NewBufferString("{}"))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), sampleErr)
	})
}

func TestCommand_SendCreateConnectionRequest(t *testing.T) {
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

		mockmsgr := &mockMessenger{MockMessenger: &mockservice.MockMessenger{}, lastID: ""}
		prov.CustomMessenger = mockmsgr

		go func() {
			for {
				if len(registrar.Services()) > 0 && mockmsgr.GetLastID() != "" { //nolint:gocritc
					replyMsg, e := service.ParseDIDCommMsgMap([]byte(fmt.Sprintf(replyMsgStr, mockmsgr.GetLastID())))
					require.NoError(t, e)

					_, e = registrar.Services()[0].HandleInbound(replyMsg, "sampleDID", "sampleTheirDID")
					require.NoError(t, e)

					break
				}
			}
		}()

		c, err := New(prov, registrar, mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		request := CreateConnectionRequest{
			Payload: json.RawMessage([]byte(fmt.Sprintf(`{"didDoc": %s}`, sampleDIDDoc))),
		}

		rqstBytes, err := json.Marshal(request)
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendCreateConnectionRequest(&b, bytes.NewBuffer(rqstBytes))
		require.NoError(t, err)
	})

	t.Run("test with empty connections", func(t *testing.T) {
		c, err := New(newMockProvider(nil), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		request := CreateConnectionRequest{
			Payload: json.RawMessage([]byte(fmt.Sprintf(`{"didDoc": %s}`, sampleDIDDoc))),
		}

		rqstBytes, err := json.Marshal(request)
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendCreateConnectionRequest(&b, bytes.NewBuffer(rqstBytes))

		require.Error(t, err)
		require.Contains(t, err.Error(), "no connection found to create invitation")
	})

	t.Run("test failure while getting connections", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				GetConnectionsErr: fmt.Errorf(sampleErr),
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		request := CreateConnectionRequest{
			Payload: json.RawMessage([]byte(fmt.Sprintf(`{"didDoc": %s}`, sampleDIDDoc))),
		}

		rqstBytes, err := json.Marshal(request)
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendCreateConnectionRequest(&b, bytes.NewBuffer(rqstBytes))

		require.Error(t, err)
		require.Contains(t, err.Error(), sampleErr)
	})

	t.Run("test failure while sending message", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				Connections: []string{"sample-connection"},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		request := CreateConnectionRequest{
			Payload: json.RawMessage([]byte(fmt.Sprintf(`{"didDoc": %s}`, sampleDIDDoc))),
		}

		rqstBytes, err := json.Marshal(request)
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendCreateConnectionRequest(&b, bytes.NewBuffer(rqstBytes))

		require.Error(t, err)
	})

	t.Run("test failure due to invalid request", func(t *testing.T) {
		prov := newMockProvider(map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{
				Connections: []string{"sample-connection"},
			},
			didexchangesvc.DIDExchange: &sdkmockprotocol.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		})

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer
		err = c.SendCreateConnectionRequest(&b, bytes.NewBufferString("---"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid character")
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

type mockMessenger struct {
	*mockservice.MockMessenger
	lastID string
	lock   sync.RWMutex
}

// Send mock messenger Send.
func (m *mockMessenger) Send(msg service.DIDCommMsgMap, myDID, theirDID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.lastID = msg.ID()

	return nil
}

// GetLastID returns ID of the last message received.
func (m *mockMessenger) GetLastID() string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.lastID
}
