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

	didexchangesvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	outofbandsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofband"
	mockmsghandler "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/msghandler"
	mockdidexchange "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockprotocol "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstorage "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/mediatorclient"
	sdkmockprotocol "github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks/protocol"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/testutil"
)

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New(newMockProvider(nil), mockmsghandler.NewMockMsgServiceProvider())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotEmpty(t, c.GetRESTHandlers())
		require.Len(t, c.GetRESTHandlers(), 1)
	})

	t.Run("test failure while creating mediator client", func(t *testing.T) {
		c, err := New(&mockprotocol.Provider{}, mockmsghandler.NewMockMsgServiceProvider())
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

		cmd, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, ConnectPath)

		buf, err := testutil.GetSuccessResponseFromHandler(handler, bytes.NewBufferString(sampleInvitation), handler.Path())
		require.NoError(t, err)

		resp := &connectionResponse{}
		err = json.NewDecoder(buf).Decode(&resp.Response)
		require.NoError(t, err)

		require.Equal(t, resp.Response.ConnectionID, sampleConnID)
		require.Equal(t, resp.Response.RoutingKeys, []string{sampleRoutingKeys})
		require.Equal(t, resp.Response.RouterEndpoint, sampleRouterEndpoint)
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

		cmd, err := New(prov, mockmsghandler.NewMockMsgServiceProvider())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, ConnectPath)

		buf, code, err := testutil.SendRequestToHandler(handler, bytes.NewBufferString(sampleInvitation), handler.Path())
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		require.Equal(t, http.StatusInternalServerError, code)
		testutil.VerifyError(t, mediatorclient.ConnectMediatorError, "sample-error", buf.Bytes())
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
