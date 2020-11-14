/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package mediatorclient // nolint:testpackage // uses internal implementation details

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexchangesvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
	mockmsghandler "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/msghandler"
	mockdidexchange "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockprotocol "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstorage "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	sdkmockprotocol "github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks/protocol"
)

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New(newMockProvider(nil), mockmsghandler.NewMockMsgServiceProvider())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotEmpty(t, c.GetHandlers())
		require.Len(t, c.GetHandlers(), 1)
	})

	t.Run("test failure while creating mediator client", func(t *testing.T) {
		c, err := New(&mockprotocol.Provider{}, mockmsghandler.NewMockMsgServiceProvider())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create mediator client")
	})

	t.Run("test failure while creating did-exchange client", func(t *testing.T) {
		c, err := New(&mockprotocol.Provider{
			ServiceMap: map[string]interface{}{
				mediatorsvc.Coordination: &mockroute.MockMediatorSvc{},
			},
		}, mockmsghandler.NewMockMsgServiceProvider())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create did-exchange client")
	})

	t.Run("test failure while creating out-of-band client", func(t *testing.T) {
		c, err := New(&mockprotocol.Provider{
			ServiceMap: map[string]interface{}{
				mediatorsvc.Coordination:   &mockroute.MockMediatorSvc{},
				didexchangesvc.DIDExchange: &mockdidexchange.MockDIDExchangeSvc{},
			},
			StorageProviderValue:              mockstorage.NewMockStoreProvider(),
			ProtocolStateStorageProviderValue: mockstorage.NewMockStoreProvider(),
		}, mockmsghandler.NewMockMsgServiceProvider())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create out-of-band client")
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
		sampleErr = "sample-error"
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
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

		c, err := New(prov, mockMsgRegistrar)
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
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

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
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
		})
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

		c, err := New(prov, &mockmsghandler.MockMsgSvcProvider{UnregisterErr: fmt.Errorf(sampleErr)})
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

func newMockProvider(serviceMap map[string]interface{}) *mockprotocol.Provider {
	if serviceMap == nil {
		serviceMap = map[string]interface{}{
			mediatorsvc.Coordination:   &mockroute.MockMediatorSvc{},
			didexchangesvc.DIDExchange: &mockdidexchange.MockDIDExchangeSvc{},
			outofbandsvc.Name:          &sdkmockprotocol.MockOobService{},
		}
	}

	return &mockprotocol.Provider{
		ServiceMap:                        serviceMap,
		StorageProviderValue:              mockstorage.NewMockStoreProvider(),
		ProtocolStateStorageProviderValue: mockstorage.NewMockStoreProvider(),
	}
}
