//go:build js && wasm
// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
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

	"github.com/trustbloc/agent-sdk/pkg/agentsetup"
	"github.com/trustbloc/agent-sdk/pkg/wasmsetup"
)

var logger = log.New("agent-js-worker")

const (
	commandPkg = "agent"
	startFn    = "Start"
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

		ctx, err := agentOpts(opts)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		handlers, err := getAriesHandlers(ctx, opts)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		// add command handlers
		wasmsetup.AddCommandHandlers(handlers, pkgMap)

		// TODO: uncomment addStopAgentHandler after proper refactoring of agent-js-worker go code.
		// addStopAgentHandler(a, pkgMap)

		return &wasmsetup.Result{
			ID:      c.ID,
			Payload: map[string]interface{}{"message": "agent started successfully"},
		}
	}

	pkgMap[commandPkg] = fnMap
}

func getAriesHandlers(ctx *walletLiteProvider,
	opts *agentsetup.AgentStartOpts) ([]controllercmd.Handler, error) {
	wallet := vcwalletcmd.New(ctx, &vcwalletcmd.Config{
		WebKMSCacheSize:                  opts.CacheSize,
		EDVReturnFullDocumentsOnQuery:    true,
		EDVBatchEndpointExtensionEnabled: true,
		WebKMSAuthzProvider:              &agentsetup.WebkmsZCAPSigner{},
		EdvAuthzProvider:                 &agentsetup.EdvZCAPSigner{},
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

type walletLiteProvider struct {
	storageProvider      storage.Provider
	vDRegistry           vdrapi.Registry
	crypto               cryptoapi.Crypto
	kms                  kms.KeyManager
	jSONLDDocumentLoader jsonld.DocumentLoader
	mediaTypeProfiles    []string
}

func (p *walletLiteProvider) StorageProvider() storage.Provider {
	return p.storageProvider
}

func (p *walletLiteProvider) VDRegistry() vdrapi.Registry {
	return p.vDRegistry
}

func (p *walletLiteProvider) KMS() kms.KeyManager {
	return p.kms
}

func (p *walletLiteProvider) Crypto() cryptoapi.Crypto {
	return p.crypto
}

func (p *walletLiteProvider) JSONLDDocumentLoader() jsonld.DocumentLoader {
	return p.jSONLDDocumentLoader
}

func (p *walletLiteProvider) MediaTypeProfiles() []string {
	return p.mediaTypeProfiles
}

func agentOpts(startOpts *agentsetup.AgentStartOpts) (*walletLiteProvider, error) {
	var options []aries.Option

	provider := &walletLiteProvider{}

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

	provider.vDRegistry = vrd

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
