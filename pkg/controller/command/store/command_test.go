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

	storeutil "github.com/hyperledger/aries-framework-go/component/storageutil/mock"
	"github.com/hyperledger/aries-framework-go/spi/storage"
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

	require.Len(t, cmd.GetHandlers(), 5)
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

func TestCommand_Query(t *testing.T) {
	t.Run("Failure during store query call", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{ErrQuery: errors.New("query failure")}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		req, err := json.Marshal(QueryRequest{Expression: "expression", PageSize: 5})
		require.NoError(t, err)

		require.EqualError(t, cmd.Query(res, bytes.NewBuffer(req)), "query failure")
	})

	t.Run("Failure while getting first result from iterator", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store =
			&mocks.MockStore{QueryReturnItr: &mocks.MockIterator{ErrNext: errors.New("next failure")}}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		req, err := json.Marshal(QueryRequest{Expression: "expression", PageSize: 5})
		require.NoError(t, err)

		require.EqualError(t, cmd.Query(res, bytes.NewBuffer(req)), "next failure")
	})

	t.Run("Empty request", func(t *testing.T) {
		cmd, err := New(&protocol.MockProvider{StoreProvider: mocks.NewMockStoreProvider()})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Query(res, bytes.NewBufferString(``)), io.EOF.Error())
	})

	t.Run("Success", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{QueryReturnItr: &mocks.MockIterator{MoreResults: true}}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		req, err := json.Marshal(QueryRequest{Expression: "expression", PageSize: 5})
		require.NoError(t, err)
		require.NoError(t, cmd.Query(res, bytes.NewBuffer(req)))

		var resp *QueryResponse
		require.NoError(t, json.Unmarshal(res.Bytes(), &resp))

		require.Len(t, resp.Results, 1)
		require.Equal(t, "Value", string(resp.Results[0]))
	})

	t.Run("pageSize 0", func(t *testing.T) {
		expression := "expression"

		storeProvider := &storeutil.Provider{
			OpenStoreReturn: &mockStore{
				queryFunc: func(e string, options ...storage.QueryOption) (storage.Iterator, error) {
					require.Equal(t, expression, e)
					require.Empty(t, options)

					return &mocks.MockIterator{MoreResults: true}, nil
				},
			},
		}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		req, err := json.Marshal(QueryRequest{Expression: expression, PageSize: 0})
		require.NoError(t, err)
		require.NoError(t, cmd.Query(res, bytes.NewBuffer(req)))

		var resp *QueryResponse
		require.NoError(t, json.Unmarshal(res.Bytes(), &resp))

		require.Len(t, resp.Results, 1)
		require.Equal(t, "Value", string(resp.Results[0]))
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

func TestCommand_Flush(t *testing.T) {
	t.Run("Failure during store flush call", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{ErrFlush: errors.New("flush failure")}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.EqualError(t, cmd.Flush(res, nil), "flush failure")
	})

	t.Run("Success", func(t *testing.T) {
		storeProvider := mocks.NewMockStoreProvider()
		storeProvider.Store = &mocks.MockStore{Store: map[string][]byte{"key": []byte(`value`)}}

		cmd, err := New(&protocol.MockProvider{StoreProvider: storeProvider})
		require.NoError(t, err)
		require.NotNil(t, cmd)

		res := &bytes.Buffer{}

		require.NoError(t, cmd.Flush(res, bytes.NewBuffer(nil)))
	})
}

type mockStore struct {
	queryFunc func(string, ...storage.QueryOption) (storage.Iterator, error)
}

func (m *mockStore) Put(key string, value []byte, tags ...storage.Tag) error {
	panic("implement me")
}

func (m *mockStore) Get(key string) ([]byte, error) {
	panic("implement me")
}

func (m *mockStore) GetTags(key string) ([]storage.Tag, error) {
	panic("implement me")
}

func (m *mockStore) GetBulk(keys ...string) ([][]byte, error) {
	panic("implement me")
}

func (m *mockStore) Query(expression string, options ...storage.QueryOption) (storage.Iterator, error) {
	return m.queryFunc(expression, options...)
}

func (m *mockStore) Delete(key string) error {
	panic("implement me")
}

func (m *mockStore) Batch(operations []storage.Operation) error {
	panic("implement me")
}

func (m *mockStore) Flush() error {
	panic("implement me")
}

func (m *mockStore) Close() error {
	panic("implement me")
}
