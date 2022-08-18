//go:build js && wasm
// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"io"
	"net/http"
	"syscall/js"

	controllercmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"
	kmscmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/kms"
	vcwalletcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/vcwallet"
	vdrcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/vdr"
	cryptoapi "github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	vdrapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/peer"
	"github.com/hyperledger/aries-framework-go/spi/storage"
	"github.com/mitchellh/mapstructure"
	jsonld "github.com/piprate/json-gold/ld"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/store"

	"github.com/trustbloc/agent-sdk/pkg/agentsetup"
	didclientcmd "github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
	"github.com/trustbloc/agent-sdk/pkg/wasmsetup"
)

var logger = log.New("agent-js-worker")

const (
	commandPkg = "agent"
	startFn    = "Start"
	stopFn     = "Stop"
	workers    = 2

	defaultEndpoint = "didcomm:transport/queue"
)

// TODO Signal JS when WASM is loaded and ready.
//      This is being used in tests for now.
var (
	ready  = make(chan struct{}) //nolint:gochecknoglobals
	isTest = false               //nolint:gochecknoglobals
)

// main registers the 'handleMsg' function in the JS context's global scope to receive commands.
// Results are posted back to the 'handleResult' JS function.
func main() {
	// TODO: capacity was added due to deadlock.
	//  Looks like js worker are not able to pick up 'output chan *wasmsetup.Result'.
	//  Another fix for that is to wrap 'in <- cmd' in a goroutine. e.g go func() { in <- cmd }()
	//  We need to figure out what is the root cause of deadlock and fix it properly.
	input := make(chan *wasmsetup.Command, 10) // nolint: gomnd
	output := make(chan *wasmsetup.Result)

	go wasmsetup.Pipe(input, output, addAgentHandlers, workers)

	go wasmsetup.SendTo(output)

	js.Global().Set("handleMsg", js.FuncOf(wasmsetup.TakeFrom(input)))

	wasmsetup.PostInitMsg()

	if isTest {
		ready <- struct{}{}
	}

	select {}
}

func addAgentHandlers(pkgMap map[string]map[string]func(*wasmsetup.Command) *wasmsetup.Result) {
	fnMap := make(map[string]func(*wasmsetup.Command) *wasmsetup.Result)
	fnMap[startFn] = func(c *wasmsetup.Command) *wasmsetup.Result {
		opts, err := startOpts(c.Payload)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		err = agentsetup.SetLogLevel(opts.LogLevel)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		services, err := createAgentServices(opts)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		ariesHandlers, err := getAriesHandlers(services, opts)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		agentHandlers, err := getAgentHandlers(services, opts)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		var handlers []controllercmd.Handler
		handlers = append(handlers, ariesHandlers...)
		handlers = append(handlers, agentHandlers...)

		// add command handlers
		wasmsetup.AddCommandHandlers(handlers, pkgMap)

		addStopAgentHandler(services, pkgMap)

		return &wasmsetup.Result{
			ID:      c.ID,
			Payload: map[string]interface{}{"message": "agent started successfully"},
		}
	}

	pkgMap[commandPkg] = fnMap
}

func getAriesHandlers(ctx *walletServices,
	opts *agentsetup.AgentStartOpts) ([]controllercmd.Handler, error) {
	var (
		headerFunc func(r2 *http.Request) (*http.Header, error)
		err        error
	)

	if opts.GNAPSigningJWK != "" && opts.GNAPAccessToken != "" {
		headerFunc, err = agentsetup.GNAPAddHeaderFunc(opts.GNAPAccessToken, opts.GNAPSigningJWK)
		if err != nil {
			return nil, fmt.Errorf("failed to create gnap header func: %w", err)
		}
	}

	wallet := vcwalletcmd.New(ctx, &vcwalletcmd.Config{
		WebKMSCacheSize:                  opts.CacheSize,
		EDVReturnFullDocumentsOnQuery:    true,
		EDVBatchEndpointExtensionEnabled: true,
		WebKMSGNAPSigner:                 headerFunc,
		EDVGNAPSigner:                    headerFunc,
	})

	vcmd, err := vdrcmd.New(ctx)
	if err != nil {
		return nil, err
	}

	kcmd := kmscmd.New(ctx)

	handlers := wallet.GetHandlers()
	handlers = append(handlers, vcmd.GetHandlers()...)
	handlers = append(handlers, kcmd.GetHandlers()...)

	return handlers, nil
}

func getAgentHandlers(ctx *walletServices,
	opts *agentsetup.AgentStartOpts) ([]controllercmd.Handler, error) {
	// did client command operation.
	didClientCmd, err := didclientcmd.New(opts.BlocDomain, opts.DidAnchorOrigin, opts.SidetreeToken,
		opts.UnanchoredDIDMaxLifeTime, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DID client: %w", err)
	}

	storeCmd, err := store.New(ctx)
	if err != nil {
		return nil, err
	}

	var handlers []controllercmd.Handler
	handlers = append(handlers, didClientCmd.GetHandlers()...)
	handlers = append(handlers, storeCmd.GetHandlers()...)

	return handlers, nil
}

func startOpts(payload map[string]interface{}) (*agentsetup.AgentStartOpts, error) {
	logger.Debugf("agent start options: %+v\n", payload)

	opts := &agentsetup.AgentStartOpts{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  opts,
	})
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(payload)
	if err != nil {
		return nil, err
	}

	if opts.UserConfig == nil {
		opts.UserConfig = &agentsetup.UserConfig{}
	}

	return opts, nil
}

func addStopAgentHandler(a io.Closer, pkgMap map[string]map[string]func(*wasmsetup.Command) *wasmsetup.Result) {
	fnMap := make(map[string]func(*wasmsetup.Command) *wasmsetup.Result)
	fnMap[stopFn] = func(c *wasmsetup.Command) *wasmsetup.Result {
		err := a.Close()
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		// reset handlers when stopped
		for k := range pkgMap {
			delete(pkgMap, k)
		}

		// put back start wasmsetup.Command once stopped
		addAgentHandlers(pkgMap)

		return &wasmsetup.Result{
			ID:      c.ID,
			Payload: map[string]interface{}{"message": "agent stopped"},
		}
	}
	pkgMap[commandPkg] = fnMap
}

type walletServices struct {
	storageProvider      storage.Provider
	vdrRegistry          vdrapi.Registry
	crypto               cryptoapi.Crypto
	kms                  kms.KeyManager
	jSONLDDocumentLoader jsonld.DocumentLoader
	mediaTypeProfiles    []string
}

func (p *walletServices) StorageProvider() storage.Provider {
	return p.storageProvider
}

func (p *walletServices) VDRegistry() vdrapi.Registry {
	return p.vdrRegistry
}

func (p *walletServices) KMS() kms.KeyManager {
	return p.kms
}

func (p *walletServices) Crypto() cryptoapi.Crypto {
	return p.crypto
}

func (p *walletServices) JSONLDDocumentLoader() jsonld.DocumentLoader {
	return p.jSONLDDocumentLoader
}

func (p *walletServices) MediaTypeProfiles() []string {
	return p.mediaTypeProfiles
}

// Close frees resources being maintained by the framework.
func (p *walletServices) Close() error {
	if p.storageProvider != nil {
		err := p.storageProvider.Close()
		if err != nil {
			return fmt.Errorf("failed to close the store: %w", err)
		}
	}

	if p.vdrRegistry != nil {
		if err := p.vdrRegistry.Close(); err != nil {
			return fmt.Errorf("vdr registry close failed: %w", err)
		}
	}

	return nil
}

func createAgentServices(startOpts *agentsetup.AgentStartOpts) (*walletServices, error) {
	var options []aries.Option

	provider := &walletServices{}

	if len(startOpts.MediaTypeProfiles) > 0 {
		provider.mediaTypeProfiles = startOpts.MediaTypeProfiles
	}

	// indexedDBProvider used by localKMS and JSON-LD contexts
	indexedDBProvider, err := agentsetup.CreateIndexedDBStorage(startOpts)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while creating IndexDB storage provider: %w", err)
	}

	provider.storageProvider = indexedDBProvider

	loader, err := agentsetup.CreateJSONLDDocumentLoader(indexedDBProvider, startOpts.ContextProviderURLs)
	if err != nil {
		return nil, fmt.Errorf("create document loader: %w", err)
	}

	provider.jSONLDDocumentLoader = loader

	var cryptoImpl cryptoapi.Crypto

	kmsImpl, cryptoImpl, _, err := agentsetup.CreateKMSAndCrypto(startOpts, indexedDBProvider, options)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while creating LocalKMS and Crypto instance: %w", err)
	}

	provider.kms = kmsImpl
	provider.crypto = cryptoImpl

	vdrs, err := agentsetup.CreateVDRs(startOpts.HTTPResolvers, startOpts.BlocDomain, startOpts.UnanchoredDIDMaxLifeTime)
	if err != nil {
		return nil, err
	}

	vrd, err := createVDR(vdrs, startOpts, provider.storageProvider)
	if err != nil {
		return nil, err
	}

	provider.vdrRegistry = vrd

	return provider, nil
}

func createVDR(vdrs []vdrapi.VDR, startOpts *agentsetup.AgentStartOpts,
	storageProvider storage.Provider) (*vdr.Registry, error) {
	var opts []vdr.Option
	for _, v := range vdrs {
		opts = append(opts, vdr.WithVDR(v))
	}

	p, err := peer.New(storageProvider)
	if err != nil {
		return nil, fmt.Errorf("create new vdr peer failed: %w", err)
	}

	dst := vdrapi.DIDCommServiceType

	for _, mediaType := range startOpts.MediaTypeProfiles {
		if mediaType == transport.MediaTypeDIDCommV2Profile || mediaType == transport.MediaTypeAIP2RFC0587Profile {
			dst = vdrapi.DIDCommV2ServiceType

			break
		}
	}

	opts = append(opts,
		vdr.WithVDR(p),
		vdr.WithDefaultServiceType(dst),
		vdr.WithDefaultServiceEndpoint(defaultEndpoint),
	)

	k := key.New()
	opts = append(opts, vdr.WithVDR(k))

	return vdr.New(opts...), nil
}
