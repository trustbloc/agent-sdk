// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package jsindexeddbcache

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messenger"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/introduce"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/store/connection"
	"github.com/hyperledger/aries-framework-go/pkg/store/did"
	"github.com/hyperledger/aries-framework-go/pkg/store/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/peer"

	"github.com/hyperledger/aries-framework-go/component/storage/jsindexeddb"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/trustbloc/edge-core/pkg/log"
)

const (
	dbName    = "aries-%s"
	defDBName = "aries"
)

var logger = log.New("jsindexeddb-cache")

// Provider jsindexdbcache implementation of storage.Provider interface.
type Provider struct {
	jsindexeddbProvider storage.Provider
	storesName          map[string]string
}

// NewProvider instantiates Provider.
func NewProvider(name string, clearCache time.Duration) (*Provider, error) {
	jsindexeddbProvider, err := jsindexeddb.NewProvider(name)
	if err != nil {
		return nil, err
	}

	db := defDBName
	if name != "" {
		db = fmt.Sprintf(dbName, name)
	}

	m := make(map[string]string)

	for _, v := range getStoreNames() {
		m[v] = db
		// TODO find way to clear cache when user close browser
		// Aries agent gets re-initialized every time CHAPI window opens
	}

	prov := &Provider{jsindexeddbProvider: jsindexeddbProvider, storesName: m}

	ticker := time.NewTicker(clearCache)
	quit := make(chan struct{})
	go func(p *Provider) {
		for {
			select {
			case <-ticker.C:
				if err := p.Close(); err != nil {
					logger.Errorf(err.Error())
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(prov)

	return prov, nil
}

// Close closes all stores created under this store provider.
func (p *Provider) Close() error {
	for storeName, databaseName := range p.storesName {
		if err := clearStore(databaseName, storeName); err != nil {
			return err
		}
	}

	return nil
}

// CloseStore close store
func (p *Provider) CloseStore(_ string) error {
	return nil
}

// OpenStore open store.
func (p *Provider) OpenStore(name string) (storage.Store, error) {
	store, err := p.jsindexeddbProvider.OpenStore(name)
	if err != nil {
		return nil, err
	}

	_, exist := p.storesName[name]
	if !exist {
		databaseName := fmt.Sprintf(dbName, name)
		p.storesName[name] = databaseName
		// TODO find way to clear cache when user close browser
		// Aries agent gets re-initialized every time CHAPI window opens
	}

	return store, nil
}

func clearStore(databaseName, storeName string) error {
	req := js.Global().Get("indexedDB").Call("open", databaseName, 1)
	v, err := getResult(req)
	if err != nil {
		return err
	}

	req = v.Call("transaction", storeName, "readwrite").Call("objectStore", storeName).Call("clear")
	_, err = getResult(req)
	if err != nil {
		return err
	}

	return nil
}

func getResult(req js.Value) (*js.Value, error) {
	onsuccess := make(chan js.Value)
	onerror := make(chan js.Value)

	const timeout = 10

	req.Set("onsuccess", js.FuncOf(func(this js.Value, inputs []js.Value) interface{} {
		onsuccess <- this.Get("result")
		return nil
	}))
	req.Set("onerror", js.FuncOf(func(this js.Value, inputs []js.Value) interface{} {
		onerror <- this.Get("error")
		return nil
	}))
	select {
	case value := <-onsuccess:
		return &value, nil
	case value := <-onerror:
		return nil, fmt.Errorf("%s %s", value.Get("name").String(),
			value.Get("message").String())
	case <-time.After(timeout * time.Second):
		return nil, fmt.Errorf("timeout waiting for event")
	}
}

func getStoreNames() []string {
	return []string{
		messenger.MessengerStore,
		mediator.Coordination,
		connection.Namespace,
		introduce.Introduce,
		peer.StoreNamespace,
		did.StoreName,
		localkms.Namespace,
		verifiable.NameSpace,
		issuecredential.Name,
		presentproof.Name,
	}
}
