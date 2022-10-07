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

	ariesctrl "github.com/hyperledger/aries-framework-go/pkg/controller"
	controllercmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"
	vcwalletcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/vcwallet"
	cryptoapi "github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/mitchellh/mapstructure"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/agentsetup"
	agentctrl "github.com/trustbloc/agent-sdk/pkg/controller"

	"github.com/trustbloc/agent-sdk/pkg/wasmsetup"
)

var logger = log.New("agent-js-worker")

const (
	commandPkg = "agent"
	startFn    = "Start"
	stopFn     = "Stop"
	workers    = 2
)

// TODO Signal JS when WASM is loaded and ready.
//      This is being used in tests for now.
var (
	ready  = make(chan struct{}) //nolint:gochecknoglobals
	isTest = false               //nolint:gochecknoglobals
)

// main registers the 'handleMsg' function in the JS context's global scope to receive commands.
// Results are posted back to the 'handleResult' JS function.
// nolint:lll
func main() {
	// TODO: capacity was added due to deadlock. Looks like js worker are not able to pick up 'output chan *wasmsetup.Result'.
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

		ariesOpts, err := ariesAgentOpts(opts)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		msgHandler := msghandler.NewRegistrar()
		ariesOpts = append(ariesOpts, aries.WithMessageServiceProvider(msgHandler))

		a, err := aries.New(ariesOpts...)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		ctx, err := a.Context()
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		handlers, err := getAriesHandlers(ctx, msgHandler, opts)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		agentHandlers, err := getAgentHandlers(ctx, msgHandler, opts)
		if err != nil {
			return wasmsetup.NewErrResult(c.ID, err.Error())
		}

		handlers = append(handlers, agentHandlers...)

		// add command handlers
		wasmsetup.AddCommandHandlers(handlers, pkgMap)

		// add stop agent handler
		addStopAgentHandler(a, pkgMap)

		return &wasmsetup.Result{
			ID:      c.ID,
			Payload: map[string]interface{}{"message": "agent started successfully"},
		}
	}

	pkgMap[commandPkg] = fnMap
}

func getAriesHandlers(ctx *context.Provider, r controllercmd.MessageHandler,
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

	handlers, err := ariesctrl.GetCommandHandlers(ctx, ariesctrl.WithMessageHandler(r),
		ariesctrl.WithDefaultLabel(opts.Label), ariesctrl.WithNotifier(&wasmsetup.JSNotifier{}),
		ariesctrl.WithWalletConfiguration(&vcwalletcmd.Config{
			WebKMSCacheSize:                  opts.CacheSize,
			EDVReturnFullDocumentsOnQuery:    true,
			EDVBatchEndpointExtensionEnabled: true,
			WebKMSGNAPSigner:                 headerFunc,
			EDVGNAPSigner:                    headerFunc,
			ValidateDataModel:                opts.ValidateDataModel,
		}))
	if err != nil {
		return nil, err
	}

	return handlers, nil
}

func getAgentHandlers(ctx *context.Provider,
	r controllercmd.MessageHandler, opts *agentsetup.AgentStartOpts) ([]controllercmd.Handler, error) {
	handlers, err := agentctrl.GetCommandHandlers(ctx, agentctrl.WithBlocDomain(opts.BlocDomain),
		agentctrl.WithDidAnchorOrigin(opts.DidAnchorOrigin), agentctrl.WithSidetreeToken(opts.SidetreeToken),
		agentctrl.WithUnanchoredDIDMaxLifeTime(opts.UnanchoredDIDMaxLifeTime), agentctrl.WithMessageHandler(r),
		agentctrl.WithNotifier(&wasmsetup.JSNotifier{}))
	if err != nil {
		return nil, err
	}

	return handlers, nil
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

//nolint:gocyclo,funlen
func ariesAgentOpts(startOpts *agentsetup.AgentStartOpts) ([]aries.Option, error) {
	var options []aries.Option

	msgHandler := msghandler.NewRegistrar()
	options = append(options, aries.WithMessageServiceProvider(msgHandler))

	if startOpts.TransportReturnRoute != "" {
		options = append(options, aries.WithTransportReturnRoute(startOpts.TransportReturnRoute))
	}

	if len(startOpts.MediaTypeProfiles) > 0 {
		options = append(options, aries.WithMediaTypeProfiles(startOpts.MediaTypeProfiles))
	}

	// indexedDBProvider used by localKMS and JSON-LD contexts
	indexedDBProvider, err := agentsetup.CreateIndexedDBStorage(startOpts)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while creating IndexDB storage provider: %w", err)
	}

	loader, err := agentsetup.CreateJSONLDDocumentLoader(indexedDBProvider, startOpts.ContextProviderURLs)
	if err != nil {
		return nil, fmt.Errorf("create document loader: %w", err)
	}

	options = append(options, aries.WithJSONLDDocumentLoader(loader))

	var (
		kmsImpl    kms.KeyManager
		cryptoImpl cryptoapi.Crypto
	)

	kmsImpl, cryptoImpl, options, err = agentsetup.CreateKMSAndCrypto(startOpts, indexedDBProvider, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create kms and crypto: %w", err)
	}

	options, err = agentsetup.AddStorageOptions(startOpts, indexedDBProvider, kmsImpl, cryptoImpl, options)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while adding storage: %w", err)
	}

	VDRs, err := agentsetup.CreateVDRs(startOpts.HTTPResolvers, startOpts.BlocDomain, startOpts.UnanchoredDIDMaxLifeTime)
	if err != nil {
		return nil, err
	}

	for i := range VDRs {
		options = append(options, aries.WithVDR(VDRs[i]))
	}

	if len(startOpts.MediaTypeProfiles) > 0 {
		options = append(options, aries.WithMediaTypeProfiles(startOpts.MediaTypeProfiles))
	}

	if len(startOpts.KeyType) > 0 {
		options = append(options, aries.WithKeyType(agentsetup.KeyTypes[startOpts.KeyType]))
	}

	if len(startOpts.KeyAgreementType) > 0 {
		options = append(options, aries.WithKeyAgreementType(agentsetup.KeyAgreementTypes[startOpts.KeyAgreementType]))
	}

	return agentsetup.AddOutboundTransports(startOpts, options)
}
