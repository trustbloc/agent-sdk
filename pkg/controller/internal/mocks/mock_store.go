/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"errors"
	"fmt"
	"sync"

	"github.com/hyperledger/aries-framework-go/spi/storage"
)

// MockStoreProvider mock store provider.
type MockStoreProvider struct {
	Store              *MockStore
	Custom             storage.Store
	ErrOpenStoreHandle error
	ErrClose           error
	ErrCloseStore      error
	FailNamespace      string
}

// NewMockStoreProvider new store provider instance.
func NewMockStoreProvider() *MockStoreProvider {
	return &MockStoreProvider{Store: &MockStore{
		Store: make(map[string][]byte),
	}}
}

// NewCustomMockStoreProvider new mock store provider instance
// from existing mock store.
func NewCustomMockStoreProvider(customStore storage.Store) *MockStoreProvider {
	return &MockStoreProvider{Custom: customStore}
}

// OpenStore opens and returns a store for given name space.
func (s *MockStoreProvider) OpenStore(name string) (storage.Store, error) {
	if name == s.FailNamespace {
		return nil, fmt.Errorf("failed to open store for name space %s", name)
	}

	if s.Custom != nil {
		return s.Custom, s.ErrOpenStoreHandle
	}

	return s.Store, s.ErrOpenStoreHandle
}

// SetStoreConfig is not implemented.
func (s *MockStoreProvider) SetStoreConfig(name string, config storage.StoreConfiguration) error {
	panic("implement me")
}

// GetStoreConfig is not implemented.
func (s *MockStoreProvider) GetStoreConfig(name string) (storage.StoreConfiguration, error) {
	panic("implement me")
}

// GetOpenStores returns the single mocked store.
func (s *MockStoreProvider) GetOpenStores() []storage.Store {
	if s.Custom != nil {
		return []storage.Store{s.Custom}
	}

	return []storage.Store{s.Store}
}

// Close closes all stores created under this store provider.
func (s *MockStoreProvider) Close() error {
	return s.ErrClose
}

// MockStore mock store.
type MockStore struct {
	Store          map[string][]byte
	lock           sync.RWMutex
	ErrPut         error
	ErrGet         error
	ErrQuery       error
	ErrDelete      error
	ErrFlush       error
	QueryReturnItr storage.Iterator
}

// Put stores the key and the record.
func (s *MockStore) Put(k string, v []byte, _ ...storage.Tag) error {
	if k == "" {
		return errors.New("key is mandatory")
	}

	if s.ErrPut != nil {
		return s.ErrPut
	}

	s.lock.Lock()
	s.Store[k] = v
	s.lock.Unlock()

	return s.ErrPut
}

// Get fetches the record based on key.
func (s *MockStore) Get(k string) ([]byte, error) {
	if s.ErrGet != nil {
		return nil, s.ErrGet
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	val, ok := s.Store[k]
	if !ok {
		return nil, storage.ErrDataNotFound
	}

	return val, s.ErrGet
}

// GetTags is not implemented.
func (s *MockStore) GetTags(string) ([]storage.Tag, error) {
	panic("implement me")
}

// GetBulk is not implemented.
func (s *MockStore) GetBulk(...string) ([][]byte, error) {
	panic("implement me")
}

// Query returns mocked data.
func (s *MockStore) Query(string, ...storage.QueryOption) (storage.Iterator, error) {
	return s.QueryReturnItr, s.ErrQuery
}

// Delete will delete record with k key.
func (s *MockStore) Delete(k string) error {
	s.lock.Lock()
	delete(s.Store, k)
	s.lock.Unlock()

	return s.ErrDelete
}

// Batch is not implemented.
func (s *MockStore) Batch([]storage.Operation) error {
	panic("implement me")
}

// Flush returns a mocked error.
func (s *MockStore) Flush() error {
	return s.ErrFlush
}

// Close is not implemented.
func (s *MockStore) Close() error {
	panic("implement me")
}

// MockIterator is a mocked implementation of storage.Iterator.
type MockIterator struct {
	MoreResults bool
	ErrNext     error
}

// Next returns mocked values.
func (m *MockIterator) Next() (bool, error) {
	if m.MoreResults {
		m.MoreResults = false

		return true, m.ErrNext
	}

	return false, m.ErrNext
}

// Key is not implemented.
func (m *MockIterator) Key() (string, error) {
	panic("implement me")
}

// Value returns a mocked value.
func (m *MockIterator) Value() ([]byte, error) {
	return []byte("Value"), nil
}

// Tags is not implemented.
func (m *MockIterator) Tags() ([]storage.Tag, error) {
	panic("implement me")
}

// TotalItems is not implemented.
func (m *MockIterator) TotalItems() (int, error) {
	return -1, errors.New("not implemented")
}

// Close always returns a nil error.
func (m *MockIterator) Close() error {
	return nil
}
