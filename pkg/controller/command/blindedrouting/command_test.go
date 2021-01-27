/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package blindedrouting // nolint:testpackage // uses internal implementation details

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	mockmsghandler "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/msghandler"
	mockstorage "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks"
	sdkmockprotocol "github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks/protocol"
)

const sampleErr = "sample-error"

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotEmpty(t, c.GetHandlers())
		require.Len(t, c.GetHandlers(), 2)
	})

	t.Run("test failure while creating messaging client", func(t *testing.T) {
		prov := newMockProvider()
		prov.StoreProvider = &mockstorage.MockStoreProvider{
			FailNamespace: "didexchange",
		}

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.Error(t, err)
		require.Nil(t, c)
		require.Contains(t, err.Error(), "failed to create messenger client")
	})
}

func TestCommand_SendDIDDocRequest(t *testing.T) {
	t.Run("test request validation", func(t *testing.T) {
		c, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendDIDDocRequest(&b, bytes.NewBufferString(`{}`))
		require.Error(t, err)
		require.Equal(t, err.Error(), errInvalidConnectionID)
	})

	t.Run("test send error", func(t *testing.T) {
		c, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendDIDDocRequest(&b, bytes.NewBufferString(`{"connectionID":"sample-conn-01"}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "data not found")
	})

	t.Run("test invalid request", func(t *testing.T) {
		c, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendDIDDocRequest(&b, bytes.NewBufferString(`}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid character")
	})

	const replyMsgStr = `{
							"@id": "123456781",
							"@type": "https://trustbloc.dev/blinded-routing/1.0/diddoc-resp",
							"~thread" : {"thid": "%s"},
							"data": {"didDoc": {"@id": "sample-did-id"}}
					}`

	t.Run("test send did doc request success", func(t *testing.T) {
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

		c, err := New(prov, registrar, mocks.NewMockNotifier())
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendDIDDocRequest(&b, bytes.NewBufferString(`{"connectionID":"sample-conn-01"}`))
		require.NoError(t, err)
		require.NotEmpty(t, b.Bytes())
	})
}

func TestCommand_SendRegisterRouteRequest(t *testing.T) {
	t.Run("test request validation", func(t *testing.T) {
		c, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendRegisterRouteRequest(&b, bytes.NewBufferString(`{}`))
		require.Error(t, err)
		require.Equal(t, err.Error(), errInvalidMessageID)
	})

	t.Run("test invalid request", func(t *testing.T) {
		c, err := New(newMockProvider(), mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendRegisterRouteRequest(&b, bytes.NewBufferString(`}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid character")
	})

	t.Run("test send error", func(t *testing.T) {
		prov := newMockProvider()
		mockMessenger := sdkmockprotocol.NewMockMessenger()
		mockMessenger.ErrReplyToNested = fmt.Errorf(sampleErr)
		prov.CustomMessenger = mockMessenger

		c, err := New(prov, mockmsghandler.NewMockMsgServiceProvider(), mocks.NewMockNotifier())
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendRegisterRouteRequest(&b,
			bytes.NewBufferString(`{"messageID":"sample-msg-01", "data": {"didDoc": {"@id": "sample-did-id"}}}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), sampleErr)
	})

	const replyMsgStr = `{
							"@id": "123456781",
							"@type": "https://trustbloc.dev/blinded-routing/1.0/egister-route-resp",
							"~thread" : {"thid": "%s"}
					}`

	t.Run("test send did doc request success", func(t *testing.T) {
		prov := newMockProvider()
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

		c, err := New(prov, registrar, mocks.NewMockNotifier())
		require.NoError(t, err)

		var b bytes.Buffer
		err = c.SendRegisterRouteRequest(&b,
			bytes.NewBufferString(`{"messageID":"sample-msg-01", "data": {"didDoc": {"@id": "sample-did-id"}}}`))
		require.NoError(t, err)
		require.NotEmpty(t, b.Bytes())
	})
}

func newMockProvider() *sdkmockprotocol.MockProvider {
	prov := sdkmockprotocol.NewMockProvider()
	prov.StoreProvider = mockstorage.NewMockStoreProvider()
	prov.ProtocolStateStoreProvider = mockstorage.NewMockStoreProvider()

	return prov
}
