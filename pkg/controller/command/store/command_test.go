/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/stretchr/testify/require"

	. "github.com/trustbloc/agent-sdk/pkg/controller/command/store"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/mocks/protocol"
)

func TestNew(t *testing.T) {
	storeProvider := mocks.NewMockStoreProvider()

	mock := &protocol.MockProvider{}
	mock.StoreProvider = storeProvider

	cmd, err := New(mock)
	require.NoError(t, err)
	require.NotNil(t, cmd)

	storeProvider.ErrOpenStoreHandle = errors.New("error")

	cmd, err = New(mock)
	require.EqualError(t, err, "error")
	require.Nil(t, cmd)
}

func TestCommand_GetHandlers(t *testing.T) {
	cmd, err := New(&protocol.MockProvider{StoreProvider: mocks.NewMockStoreProvider()})
	require.NoError(t, err)
	require.NotNil(t, cmd)

	require.Len(t, cmd.GetHandlers(), 4)
}

func TestCommand_Get(t *testing.T) {
	t.Run("No data", func(t *testing.T) {
		cmd, err := New(&protocol.MockProvider{StoreProvider: mocks.NewMockStoreProvider()})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Get(res, bytes.NewBufferString(`{}`)), storage.ErrDataNotFound.Error())
	})

	t.Run("Empty request", func(t *testing.T) {
		cmd, err := New(&protocol.MockProvider{StoreProvider: mocks.NewMockStoreProvider()})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Get(res, bytes.NewBufferString(``)), io.EOF.Error())
	})

	t.Run("Success", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{Store: map[string][]byte{"key": []byte(`value`)}}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		req, err := json.Marshal(GetRequest{Key: "key"})
		require.NoError(t, err)
		require.NoError(t, cmd.Get(res, bytes.NewBuffer(req)))

		var resp *GetResponse
		require.NoError(t, json.Unmarshal(res.Bytes(), &resp))

		require.Equal(t, []byte(`value`), resp.Result)
	})
}

func TestCommand_Put(t *testing.T) {
	t.Run("No data", func(t *testing.T) {
		cmd, err := New(&protocol.MockProvider{StoreProvider: mocks.NewMockStoreProvider()})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Put(res, bytes.NewBufferString(`{}`)), "key is mandatory")
	})

	t.Run("Empty request", func(t *testing.T) {
		cmd, err := New(&protocol.MockProvider{StoreProvider: mocks.NewMockStoreProvider()})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Put(res, bytes.NewBufferString(``)), io.EOF.Error())
	})

	t.Run("Success", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{Store: map[string][]byte{}}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		req, err := json.Marshal(PutRequest{Key: "key", Value: []byte(`value`)})
		require.NoError(t, err)
		require.NoError(t, cmd.Put(res, bytes.NewBuffer(req)))

		require.Equal(t, []byte(`value`), storeProvider.Store.Store["key"])
	})
}

func TestCommand_Delete(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{ErrDelete: errors.New("error")}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Delete(res, bytes.NewBufferString(`{}`)), "error")
	})

	t.Run("Empty request", func(t *testing.T) {
		cmd, err := New(&protocol.MockProvider{StoreProvider: mocks.NewMockStoreProvider()})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Delete(res, bytes.NewBufferString(``)), io.EOF.Error())
	})

	t.Run("Success", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{Store: map[string][]byte{"key": []byte(`value`)}}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		req, err := json.Marshal(DeleteRequest{Key: "key"})
		require.NoError(t, err)
		require.NoError(t, cmd.Delete(res, bytes.NewBuffer(req)))

		require.Equal(t, []byte(nil), storeProvider.Store.Store["key"])
	})
}

func TestCommand_Iterator(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{ErrItr: errors.New("error")}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Iterator(res, bytes.NewBufferString(`{}`)), "error")
	})

	t.Run("Empty request", func(t *testing.T) {
		cmd, err := New(&protocol.MockProvider{StoreProvider: mocks.NewMockStoreProvider()})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Iterator(res, bytes.NewBufferString(``)), io.EOF.Error())
	})

	t.Run("Success", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{Store: map[string][]byte{}}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		req, err := json.Marshal(IteratorRequest{})
		require.NoError(t, err)
		require.NoError(t, cmd.Iterator(res, bytes.NewBuffer(req)))
	})
}
