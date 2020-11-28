// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package jsindexeddbcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const sampleDBName = "testdb"

func TestProvider(t *testing.T) {
	t.Run("Test provider with db name", func(t *testing.T) {
		prov, err := NewProvider(sampleDBName, 5*time.Minute)
		require.NoError(t, err)
		require.NotNil(t, prov)
	})
}

func TestClear(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		prov, err := NewProvider(sampleDBName, 2*time.Second)
		require.NoError(t, err)
		store, err := prov.OpenStore("test")
		require.NoError(t, err)

		const key = "did:example:123"
		data := []byte("value")

		err = store.Put(key, data)
		require.NoError(t, err)

		doc, err := store.Get(key)
		require.NoError(t, err)
		require.NotEmpty(t, doc)
		require.Equal(t, data, doc)

		time.Sleep(3 * time.Second)

		_, err = store.Get(key)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data not found")
	})
}
