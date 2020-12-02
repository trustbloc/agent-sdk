/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blindedrouting // nolint:testpackage // uses internal implementation details

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	mockmsghandler "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/msghandler"
	mockstorage "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/blindedrouting"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks"
	sdkmockprotocol "github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks/protocol"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/testutil"
)

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotEmpty(t, c.GetRESTHandlers())
		require.Len(t, c.GetRESTHandlers(), 2)
	})

	t.Run("test failure while creating mediator client", func(t *testing.T) {
		prov := newMockProvider()
		prov.StoreProvider = &mockstorage.MockStoreProvider{
			FailNamespace: "didexchange",
		}

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to initialize blinded routing command")
	})
}

func TestOperation_SendDIDDocRequest(t *testing.T) {
	const replyMsgStr = `{
							"@id": "123456781",
							"@type": "https://trustbloc.dev/blinded-routing/1.0/diddoc-resp",
							"~thread" : {"thid": "%s"},
							"data": {"didDoc": {"@id": "sample-did-id"}}
					}`

	t.Run("test success", func(t *testing.T) {
		prov := newMockProvider()

		record := &connection.Record{
			ConnectionID: "sample-conn-01",
			State:        "completed", MyDID: "mydid", TheirDID: "theirDID-001",
		}
		mockStore := &mockstorage.MockStore{Store: make(map[string][]byte)}

		connBytes, err := json.Marshal(record)
		require.NoError(t, err)
		require.NoError(t, mockStore.Put("conn_sample-conn-01", connBytes))
		prov.StoreProvider = mockstorage.NewCustomMockStoreProvider(mockStore)

		registrar := mockmsghandler.NewMockMsgServiceProvider()
		mockMessenger := sdkmockprotocol.NewMockMessenger()
		prov.CustomMessenger = mockMessenger

		go func() {
			for {
				if len(registrar.Services()) > 0 && mockMessenger.GetLastID() != "" { //nolint: gocritic
					replyMsg, e := service.ParseDIDCommMsgMap([]byte(fmt.Sprintf(replyMsgStr, mockMessenger.GetLastID())))
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

		handler := testutil.LookupHandler(t, cmd, SendDIDDocRequestPath)

		request := blindedrouting.DIDDocRequest{
			ConnectionID: "sample-conn-01",
		}

		rqstBytes, err := json.Marshal(request)
		require.NoError(t, err)

		b, err := testutil.GetSuccessResponseFromHandler(handler, bytes.NewBuffer(rqstBytes), handler.Path())
		require.NoError(t, err)

		res := didDocResponse{}
		err = json.Unmarshal(b.Bytes(), &res.Response)
		require.NoError(t, err)
		require.NotEmpty(t, res.Response)
		require.NotEmpty(t, res.Response.Payload)
	})

	t.Run("test failure", func(t *testing.T) {
		cmd, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, SendDIDDocRequestPath)

		buf, code, err := testutil.SendRequestToHandler(handler, bytes.NewBufferString("---"), handler.Path())
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		require.Equal(t, http.StatusBadRequest, code)
		testutil.VerifyError(t, blindedrouting.InvalidRequestErrorCode, "invalid character", buf.Bytes())
	})
}

func TestOperation_SendRegisterRouteRequest(t *testing.T) {
	const replyMsgStr = `{
							"@id": "123456781",
							"@type": "https://trustbloc.dev/blinded-routing/1.0/egister-route-resp",
							"~thread" : {"thid": "%s"}
					}`

	t.Run("test success", func(t *testing.T) {
		prov := newMockProvider()
		registrar := mockmsghandler.NewMockMsgServiceProvider()
		mockMessenger := sdkmockprotocol.NewMockMessenger()
		prov.CustomMessenger = mockMessenger

		go func() {
			for {
				if len(registrar.Services()) > 0 && mockMessenger.GetLastID() != "" { //nolint:gocritc
					replyMsg, e := service.ParseDIDCommMsgMap([]byte(fmt.Sprintf(replyMsgStr, mockMessenger.GetLastID())))
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

		handler := testutil.LookupHandler(t, cmd, SendRegisterRouteRequest)

		request := blindedrouting.RegisterRouteRequest{
			MessageID:   "sample-conn-01",
			DIDDocument: json.RawMessage([]byte(`{"id": "did:example:21tDAKCER"}`)),
		}

		rqstBytes, err := json.Marshal(request)
		require.NoError(t, err)

		b, err := testutil.GetSuccessResponseFromHandler(handler, bytes.NewBuffer(rqstBytes), handler.Path())
		require.NoError(t, err)

		res := registerRouteResponse{}
		err = json.Unmarshal(b.Bytes(), &res.Response)
		require.NoError(t, err)
		require.NotEmpty(t, res.Response)
		require.NotEmpty(t, res.Response.Payload)
	})

	t.Run("test failure", func(t *testing.T) {
		cmd, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, cmd)

		handler := testutil.LookupHandler(t, cmd, SendRegisterRouteRequest)

		buf, code, err := testutil.SendRequestToHandler(handler, bytes.NewBufferString("---"), handler.Path())
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		require.Equal(t, http.StatusBadRequest, code)
		testutil.VerifyError(t, blindedrouting.InvalidRequestErrorCode, "invalid character", buf.Bytes())
	})
}

func newMockProvider() *sdkmockprotocol.MockProvider {
	prov := sdkmockprotocol.NewMockProvider()
	prov.StoreProvider = mockstorage.NewMockStoreProvider()
	prov.ProtocolStateStoreProvider = mockstorage.NewMockStoreProvider()

	return prov
}
