// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"syscall/js"
	"time"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/subtle/random"
	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/component/storage/jsindexeddb"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	ariesctrl "github.com/hyperledger/aries-framework-go/pkg/controller"
	controllercmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"
	cryptoapi "github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto/primitive/composite/keyio"
	webcrypto "github.com/hyperledger/aries-framework-go/pkg/crypto/webkms"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	arieshttp "github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/http"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/webkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage/edv"
	"github.com/hyperledger/aries-framework-go/pkg/storage/formattedstore"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/httpbinding"
	"github.com/mitchellh/mapstructure"
	"github.com/trustbloc/edge-core/pkg/log"
	kmszcap "github.com/trustbloc/hub-kms/pkg/restapi/kms/operation"
	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc"

	"github.com/trustbloc/agent-sdk/pkg/auth/zcapld"
	agentctrl "github.com/trustbloc/agent-sdk/pkg/controller"
	"github.com/trustbloc/agent-sdk/pkg/storage/jsindexeddbcache"
)

var logger = log.New("agent-js-worker")

const (
	wasmStartupTopic         = "asset-ready"
	handleResultFn           = "handleResult"
	commandPkg               = "agent"
	startFn                  = "Start"
	stopFn                   = "Stop"
	workers                  = 2
	storageTypeIndexedDB     = "indexedDB"
	storageTypeEDV           = "edv"
	validStorageTypesMsg     = "Valid storage types: " + storageTypeEDV + ", " + storageTypeIndexedDB
	blankStorageTypeErrMsg   = "no storage type specified. " + validStorageTypesMsg
	invalidStorageTypeErrMsg = "%s is not a valid storage type. " + validStorageTypesMsg
	kmsTypeWebKMS            = "webkms"
	hmacKeyIDDBKeyName       = "hmackeyid"
	keyIDStoreName           = "keyid"
	ecdhesKeyIDDBKeyName     = "ecdheskeyid"
	masterKeyStoreName       = "MasterKey"
	masterKeyDBKeyName       = masterKeyStoreName
	masterKeyNumBytes        = 32
	defaultClearCache        = "5m"
)

// TODO Signal JS when WASM is loaded and ready.
//      This is being used in tests for now.
var (
	ready  = make(chan struct{}) //nolint:gochecknoglobals
	isTest = false               //nolint:gochecknoglobals
)

// command is received from JS.
type command struct {
	ID      string                 `json:"id"`
	Pkg     string                 `json:"pkg"`
	Fn      string                 `json:"fn"`
	Payload map[string]interface{} `json:"payload"`
}

// result is sent back to JS.
type result struct {
	ID      string                 `json:"id"`
	IsErr   bool                   `json:"isErr"`
	ErrMsg  string                 `json:"errMsg"`
	Payload map[string]interface{} `json:"payload"`
	Topic   string                 `json:"topic"`
}

// agentStartOpts contains opts for starting agent.
type agentStartOpts struct {
	Label                string      `json:"agent-default-label"`
	HTTPResolvers        []string    `json:"http-resolver-url"`
	AutoAccept           bool        `json:"auto-accept"`
	OutboundTransport    []string    `json:"outbound-transport"`
	TransportReturnRoute string      `json:"transport-return-route"`
	LogLevel             string      `json:"log-level"`
	StorageType          string      `json:"storageType"`
	IndexedDBNamespace   string      `json:"indexedDB-namespace"`
	EDVServerURL         string      `json:"edvServerURL"`
	EDVVaultID           string      `json:"edvVaultID"`
	EDVCapability        string      `json:"edvCapability,omitempty"`
	BlocDomain           string      `json:"blocDomain"`
	TrustblocResolver    string      `json:"trustbloc-resolver"`
	AuthzKeyStoreURL     string      `json:"authzKeyStoreURL,omitempty"`
	OpsKeyStoreURL       string      `json:"opsKeyStoreURL,omitempty"`
	EDVOpsKIDURL         string      `json:"edvOpsKIDURL,omitempty"`
	EDVHMACKIDURL        string      `json:"edvHMACKIDURL,omitempty"`
	KMSType              string      `json:"kmsType"`
	UserConfig           *userConfig `json:"userConfig,omitempty"`
	UseEDVCache          bool        `json:"useEDVCache"`
	EDVClearCache        string      `json:"edvClearCache"`
	UseEDVBatch          bool        `json:"useEDVBatch"`
	EDVBatchSize         int         `json:"edvBatchSize"`
	CacheSize            int         `json:"cacheSize"`
	OPSKMSCapability     string      `json:"opsKMSCapability,omitempty"` // TODO should remove this
}

type userConfig struct {
	AccessToken string `json:"accessToken,omitempty"` // TODO should remove this
	SecretShare string `json:"walletSecretShare"`
}

type kmsProvider struct {
	storageProvider   storage.Provider
	secretLockService secretlock.Service
}

func (k kmsProvider) StorageProvider() storage.Provider {
	return k.storageProvider
}

func (k kmsProvider) SecretLock() secretlock.Service {
	return k.secretLockService
}

// main registers the 'handleMsg' function in the JS context's global scope to receive commands.
// Results are posted back to the 'handleResult' JS function.
func main() {
	// TODO: capacity was added due to deadlock. Looks like js worker are not able to pick up 'output chan *result'.
	//  Another fix for that is to wrap 'in <- cmd' in a goroutine. e.g go func() { in <- cmd }()
	//  We need to figure out what is the root cause of deadlock and fix it properly.
	input := make(chan *command, 10) // nolint: gomnd
	output := make(chan *result)

	go pipe(input, output)

	go sendTo(output)

	js.Global().Set("handleMsg", js.FuncOf(takeFrom(input)))

	postInitMsg()

	if isTest {
		ready <- struct{}{}
	}

	select {}
}

func takeFrom(in chan *command) func(js.Value, []js.Value) interface{} {
	return func(_ js.Value, args []js.Value) interface{} {
		cmd := &command{}
		if err := json.Unmarshal([]byte(args[0].String()), cmd); err != nil {
			logger.Errorf("agent wasm: unable to unmarshal input=%s. err=%s", args[0].String(), err)

			return nil
		}

		in <- cmd

		return nil
	}
}

func pipe(input chan *command, output chan *result) {
	handlers := testHandlers()

	addAgentHandlers(handlers)

	for w := 0; w < workers; w++ {
		go worker(input, output, handlers)
	}
}

func worker(input chan *command, output chan *result, handlers map[string]map[string]func(*command) *result) {
	for c := range input {
		if c.ID == "" {
			logger.Warnf("agent wasm: missing ID for input: %v", c)
		}

		if pkg, found := handlers[c.Pkg]; found {
			if fn, found := pkg[c.Fn]; found {
				output <- fn(c)

				continue
			}
		}

		output <- handlerNotFoundErr(c)
	}
}

func sendTo(out chan *result) {
	for r := range out {
		out, err := json.Marshal(r)
		if err != nil {
			logger.Errorf("agent wasm: failed to marshal response for id=%s err=%s ", r.ID, err)
		}

		js.Global().Call(handleResultFn, string(out))
	}
}

func testHandlers() map[string]map[string]func(*command) *result {
	return map[string]map[string]func(*command) *result{
		"test": {
			"echo": func(c *command) *result {
				return &result{
					ID:      c.ID,
					Payload: map[string]interface{}{"echo": c.Payload},
				}
			},
			"throwError": func(c *command) *result {
				return newErrResult(c.ID, "an error !!")
			},
			"timeout": func(c *command) *result {
				const echoTimeout = 10 * time.Second

				time.Sleep(echoTimeout)

				return &result{
					ID:      c.ID,
					Payload: map[string]interface{}{"echo": c.Payload},
				}
			},
		},
	}
}

func isStartCommand(c *command) bool {
	return c.Pkg == commandPkg && c.Fn == startFn
}

func isStopCommand(c *command) bool {
	return c.Pkg == commandPkg && c.Fn == stopFn
}

func handlerNotFoundErr(c *command) *result {
	if isStartCommand(c) {
		return newErrResult(c.ID, "Agent already started")
	} else if isStopCommand(c) {
		return newErrResult(c.ID, "Agent not running")
	}

	return newErrResult(c.ID, fmt.Sprintf("invalid pkg/fn: %s/%s, make sure agent is started", c.Pkg, c.Fn))
}

func addAgentHandlers(pkgMap map[string]map[string]func(*command) *result) {
	fnMap := make(map[string]func(*command) *result)
	fnMap[startFn] = func(c *command) *result {
		cOpts, err := startOpts(c.Payload)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		err = setLogLevel(cOpts.LogLevel)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		options, err := agentOpts(cOpts)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		msgHandler := msghandler.NewRegistrar()
		options = append(options, aries.WithMessageServiceProvider(msgHandler))

		a, err := aries.New(options...)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		ctx, err := a.Context()
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		handlers, err := getAriesHandlers(ctx, msgHandler, cOpts)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		agentHandlers, err := getAgentHandlers(ctx, msgHandler, cOpts)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		handlers = append(handlers, agentHandlers...)

		// add command handlers
		addCommandHandlers(handlers, pkgMap)

		// add stop agent handler
		addStopAgentHandler(a, pkgMap)

		return &result{
			ID:      c.ID,
			Payload: map[string]interface{}{"message": "agent started successfully"},
		}
	}

	pkgMap[commandPkg] = fnMap
}

type execFn func(rw io.Writer, req io.Reader) error

type commandHandler struct {
	name   string
	method string
	exec   execFn
}

func getAriesHandlers(ctx *context.Provider, r controllercmd.MessageHandler,
	opts *agentStartOpts) ([]commandHandler, error) {
	handlers, err := ariesctrl.GetCommandHandlers(ctx, ariesctrl.WithMessageHandler(r),
		ariesctrl.WithDefaultLabel(opts.Label), ariesctrl.WithNotifier(&jsNotifier{}))
	if err != nil {
		return nil, err
	}

	var hh []commandHandler

	for _, h := range handlers {
		handle := h.Handle()

		hh = append(hh, commandHandler{
			name:   h.Name(),
			method: h.Method(),
			exec: func(rw io.Writer, req io.Reader) error {
				e := handle(rw, req)
				if e != nil {
					return fmt.Errorf("code: %+v, message: %s", e.Code(), e.Error())
				}

				return nil
			},
		})
	}

	return hh, nil
}

func getAgentHandlers(ctx *context.Provider,
	r controllercmd.MessageHandler, opts *agentStartOpts) ([]commandHandler, error) {
	handlers, err := agentctrl.GetCommandHandlers(ctx, agentctrl.WithBlocDomain(opts.BlocDomain),
		agentctrl.WithMessageHandler(r), agentctrl.WithNotifier(&jsNotifier{}))
	if err != nil {
		return nil, err
	}

	var hh []commandHandler

	for _, h := range handlers {
		handle := h.Handle()

		hh = append(hh, commandHandler{
			name:   h.Name(),
			method: h.Method(),
			exec: func(rw io.Writer, req io.Reader) error {
				e := handle(rw, req)
				if e != nil {
					return fmt.Errorf("code: %+v, message: %s", e.Code(), e.Error())
				}

				return nil
			},
		})
	}

	return hh, nil
}

func addCommandHandlers(handlers []commandHandler, pkgMap map[string]map[string]func(*command) *result) {
	for _, h := range handlers {
		fnMap, ok := pkgMap[h.name]
		if !ok {
			fnMap = make(map[string]func(*command) *result)
		}

		fnMap[h.method] = cmdExecToFn(h.exec)
		pkgMap[h.name] = fnMap
	}
}

func cmdExecToFn(exec execFn) func(*command) *result {
	return func(c *command) *result {
		b, er := json.Marshal(c.Payload)
		if er != nil {
			return &result{
				ID:     c.ID,
				IsErr:  true,
				ErrMsg: fmt.Sprintf("agent wasm: failed to unmarshal payload. err=%s", er),
			}
		}

		req := bytes.NewBuffer(b)

		var buf bytes.Buffer

		err := exec(&buf, req)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		payload := make(map[string]interface{})

		if len(buf.Bytes()) > 0 {
			if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
				return &result{
					ID:     c.ID,
					IsErr:  true,
					ErrMsg: fmt.Sprintf("agent wasm: failed to unmarshal command result=%+v err=%s", buf.String(), err),
				}
			}
		}

		return &result{
			ID:      c.ID,
			Payload: payload,
		}
	}
}

func addStopAgentHandler(a io.Closer, pkgMap map[string]map[string]func(*command) *result) {
	fnMap := make(map[string]func(*command) *result)
	fnMap[stopFn] = func(c *command) *result {
		err := a.Close()
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		// reset handlers when stopped
		for k := range pkgMap {
			delete(pkgMap, k)
		}

		// put back start command once stopped
		addAgentHandlers(pkgMap)

		return &result{
			ID:      c.ID,
			Payload: map[string]interface{}{"message": "agent stopped"},
		}
	}
	pkgMap[commandPkg] = fnMap
}

func newErrResult(id, msg string) *result {
	return &result{
		ID:     id,
		IsErr:  true,
		ErrMsg: "agent wasm: " + msg,
	}
}

func startOpts(payload map[string]interface{}) (*agentStartOpts, error) {
	opts := &agentStartOpts{}

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
		opts.UserConfig = &userConfig{}
	}

	return opts, nil
}

func createVDRs(resolvers []string, trustblocDomain, trustblocResolver string) ([]vdr.VDR, error) {
	const numPartsResolverOption = 2
	// set maps resolver to its methods
	// e.g the set of ["trustbloc@http://resolver.com", "v1@http://resolver.com"] will be
	// {"http://resolver.com": {"trustbloc":{}, "v1":{} }}
	set := make(map[string]map[string]struct{})
	// order maps URL to its initial index
	order := make(map[string]int)

	idx := -1

	for _, resolver := range resolvers {
		r := strings.Split(resolver, "@")
		if len(r) != numPartsResolverOption {
			return nil, fmt.Errorf("invalid http resolver options found: %s", resolver)
		}

		if set[r[1]] == nil {
			set[r[1]] = map[string]struct{}{}
			idx++
		}

		order[r[1]] = idx

		set[r[1]][r[0]] = struct{}{}
	}

	VDRs := make([]vdr.VDR, len(set), len(set)+1)

	for url := range set {
		methods := set[url]

		resolverVDR, err := httpbinding.New(url, httpbinding.WithAccept(func(method string) bool {
			_, ok := methods[method]

			return ok
		}))
		if err != nil {
			return nil, fmt.Errorf("failed to create new universal resolver vdr: %w", err)
		}

		VDRs[order[url]] = resolverVDR
	}

	VDRs = append(VDRs, trustbloc.New(
		trustbloc.WithDomain(trustblocDomain),
		trustbloc.WithResolverURL(trustblocResolver),
	))

	return VDRs, nil
}

func agentOpts(startOpts *agentStartOpts) ([]aries.Option, error) {
	msgHandler := msghandler.NewRegistrar()

	var options []aries.Option
	options = append(options, aries.WithMessageServiceProvider(msgHandler))

	if startOpts.TransportReturnRoute != "" {
		options = append(options, aries.WithTransportReturnRoute(startOpts.TransportReturnRoute))
	}

	// indexedDBProvider used by localKMS only.
	indexedDBProvider, err := createIndexedDBStorage(startOpts)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while creating IndexDB storage provider: %w", err)
	}

	var (
		kmsImpl    kms.KeyManager
		cryptoImpl cryptoapi.Crypto
	)

	kmsImpl, cryptoImpl, options, err = createKMSAndCrypto(startOpts, indexedDBProvider, options)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while creating LocalKMS and Crypto instance: %w", err)
	}

	options, err = addStorageOptions(startOpts, indexedDBProvider, kmsImpl, cryptoImpl, options)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while adding storage: %w", err)
	}

	VDRs, err := createVDRs(startOpts.HTTPResolvers, startOpts.BlocDomain, startOpts.TrustblocResolver)
	if err != nil {
		return nil, err
	}

	for i := range VDRs {
		options = append(options, aries.WithVDR(VDRs[i]))
	}

	return addOutboundTransports(startOpts, options)
}

func addOutboundTransports(startOpts *agentStartOpts, options []aries.Option) ([]aries.Option, error) {
	for _, transport := range startOpts.OutboundTransport {
		switch transport {
		case "http":
			outbound, err := arieshttp.NewOutbound(arieshttp.WithOutboundHTTPClient(&http.Client{}))
			if err != nil {
				return nil, err
			}

			options = append(options, aries.WithOutboundTransports(outbound))
		case "ws":
			options = append(options, aries.WithOutboundTransports(ws.NewOutbound()))
		default:
			return nil, fmt.Errorf("unsupported transport : %s", transport)
		}
	}

	return options, nil
}

func createIndexedDBStorage(opts *agentStartOpts) (*jsindexeddb.Provider, error) {
	indexedDBKMSProvider, err := jsindexeddb.NewProvider(opts.IndexedDBNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create IndexedDB provider: %w", err)
	}

	return indexedDBKMSProvider, nil
}

func addStorageOptions(startOpts *agentStartOpts, indexedDBProvider *jsindexeddb.Provider,
	ariesKMS kms.KeyManager, ariesCrypto cryptoapi.Crypto, allAriesOptions []aries.Option) ([]aries.Option, error) {
	if startOpts.StorageType == "" {
		return nil, errors.New(blankStorageTypeErrMsg)
	}

	var store storage.Provider

	var err error

	switch startOpts.StorageType {
	case storageTypeEDV:
		store, err = createEDVStorage(startOpts, indexedDBProvider, ariesKMS, ariesCrypto)
		if err != nil {
			return nil, fmt.Errorf("failed to create storage: %w", err)
		}

		allAriesOptions = append(allAriesOptions, aries.WithProtocolStateStoreProvider(indexedDBProvider))
	case storageTypeIndexedDB:
		store = indexedDBProvider
	default:
		return nil, fmt.Errorf(invalidStorageTypeErrMsg, startOpts.StorageType)
	}

	allAriesOptions = append(allAriesOptions, aries.WithStoreProvider(store))

	return allAriesOptions, nil
}

func createEDVStorage(opts *agentStartOpts, indexedDBProvider *jsindexeddb.Provider,
	kmsImpl kms.KeyManager, cryptoImpl cryptoapi.Crypto) (storage.Provider, error) {
	store, err := createEDVProvider(opts, indexedDBProvider, kmsImpl, cryptoImpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create EDV provider: %w", err)
	}

	return store, nil
}

func createKMSAndCrypto(opts *agentStartOpts, indexedDBKMSProvider storage.Provider,
	allAriesOptions []aries.Option) (kms.KeyManager, cryptoapi.Crypto, []aries.Option, error) {
	if opts.KMSType == kmsTypeWebKMS {
		return createWebkms(opts, allAriesOptions)
	}

	return createLocalKMSAndCrypto(indexedDBKMSProvider, allAriesOptions)
}

func createWebkms(opts *agentStartOpts,
	allAriesOptions []aries.Option) (*webkms.RemoteKMS, *webcrypto.RemoteCrypto, []aries.Option, error) {
	zcapSVC := zcapld.New(opts.AuthzKeyStoreURL, opts.UserConfig.AccessToken, opts.UserConfig.SecretShare)

	httpClient := &http.Client{}

	capability, err := decodeAndGunzip(opts.OPSKMSCapability)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to prepare OPS KMS capability for use: %w", err)
	}

	wKMS := webkms.New(opts.OpsKeyStoreURL, httpClient,
		webkms.WithHeaders(func(req *http.Request) (*http.Header, error) {
			if len(capability) != 0 {
				invocationAction, err := kmszcap.CapabilityInvocationAction(req)
				if err != nil {
					return nil, fmt.Errorf("webkms: failed to determine the capability's invocation action: %w", err)
				}

				return zcapSVC.SignHeader(req, capability, invocationAction)
			}

			return nil, nil
		}))

	allAriesOptions = append(allAriesOptions,
		aries.WithKMS(func(ctx kms.Provider) (kms.KeyManager, error) {
			return wKMS, nil
		}))

	var opt []webkms.Opt
	if opts.CacheSize >= 0 {
		opt = append(opt, webkms.WithCache(opts.CacheSize))
	}

	opt = append(opt, webkms.WithHeaders(func(req *http.Request) (*http.Header, error) {
		if len(capability) != 0 {
			invocationAction, err := kmszcap.CapabilityInvocationAction(req)
			if err != nil {
				return nil, fmt.Errorf("webcrypto: failed to determine the capability's invocation action: %w", err)
			}

			return zcapSVC.SignHeader(req, capability, invocationAction)
		}

		return nil, nil
	}))

	wCrypto := webcrypto.New(opts.OpsKeyStoreURL, httpClient, opt...)

	allAriesOptions = append(allAriesOptions, aries.WithCrypto(wCrypto))

	return wKMS, wCrypto, allAriesOptions, nil
}

func decodeAndGunzip(zcap string) ([]byte, error) {
	decoded, err := base64.URLEncoding.DecodeString(zcap)
	if err != nil {
		return nil, fmt.Errorf("failed to base64URL-decode zcap: %w", err)
	}

	compressed, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip reader: %w", err)
	}

	uncompressed, err := ioutil.ReadAll(compressed)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip zcap: %w", err)
	}

	return uncompressed, nil
}

func createLocalKMSAndCrypto(indexedDBKMSProvider storage.Provider,
	allAriesOptions []aries.Option) (*localkms.LocalKMS, *tinkcrypto.Crypto, []aries.Option, error) {
	masterKeyReader, err := prepareMasterKeyReader(indexedDBKMSProvider)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to prepare master key reader: %w", err)
	}

	secretLockService, err := local.NewService(masterKeyReader, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create secret lock service: %w", err)
	}

	kmsProv := kmsProvider{
		storageProvider:   indexedDBKMSProvider,
		secretLockService: secretLockService,
	}

	localKMS, err := localkms.New("local-lock://agentSDK", &kmsProv)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create local KMS: %w", err)
	}

	if localKMS != nil {
		allAriesOptions = append(allAriesOptions,
			aries.WithKMS(func(ctx kms.Provider) (kms.KeyManager, error) {
				return localKMS, nil
			}))
	}

	c, err := tinkcrypto.New()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create local Crypto: %w", err)
	}

	allAriesOptions = append(allAriesOptions, aries.WithCrypto(c))

	return localKMS, c, allAriesOptions, nil
}

func createEDVProvider(opts *agentStartOpts, indexedDBKMSProvider *jsindexeddb.Provider, kmsImpl kms.KeyManager,
	cryptoImpl cryptoapi.Crypto) (storage.Provider, error) {
	edvProvider, err := createEDVStorageProvider(opts, indexedDBKMSProvider, kmsImpl, cryptoImpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create EDV provider: %w", err)
	}

	return edvProvider, nil
}

// prepareMasterKeyReader prepares a master key reader for secret lock usage.
func prepareMasterKeyReader(kmsSecretsStoreProvider storage.Provider) (*bytes.Reader, error) {
	masterKeyStore, err := kmsSecretsStoreProvider.OpenStore(masterKeyStoreName)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key store: %w", err)
	}

	masterKey, err := masterKeyStore.Get(masterKeyDBKeyName)
	if err != nil {
		if errors.Is(err, storage.ErrDataNotFound) {
			logger.Infof("No existing master key under store %s with ID %s was found.",
				masterKeyStoreName, masterKeyDBKeyName)

			masterKeyRaw := random.GetRandomBytes(uint32(masterKeyNumBytes))
			masterKey = []byte(base64.URLEncoding.EncodeToString(masterKeyRaw))

			putErr := masterKeyStore.Put(masterKeyDBKeyName, masterKey)
			if putErr != nil {
				return nil, fmt.Errorf("failed to put newly created master key into master key store: %w", putErr)
			}

			logger.Infof("Created a new master key under store %s with ID %s.", masterKeyStoreName, masterKeyDBKeyName)

			return bytes.NewReader(masterKey), nil
		}

		return nil, fmt.Errorf("failed to get master key from master key store: %w", err)
	}

	logger.Infof("Found an existing master key under store %s with ID %s", masterKeyStoreName, masterKeyDBKeyName)

	return bytes.NewReader(masterKey), nil
}

func createEDVStorageProvider(opts *agentStartOpts, storageProvider storage.Provider, kmsImpl kms.KeyManager,
	cryptoImpl cryptoapi.Crypto) (storage.Provider, error) {
	macCrypto, err := prepareMACCrypto(opts, storageProvider, kmsImpl, cryptoImpl)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare MAC crypto: %w", err)
	}

	edvRESTProvider, err := prepareEDVRESTProvider(opts, macCrypto)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare EDV REST provider: %w", err)
	}

	formattedProvider, err := prepareFormattedProvider(opts, storageProvider, kmsImpl, cryptoImpl, macCrypto,
		edvRESTProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare formatted provider: %w", err)
	}

	return formattedProvider, nil
}

func prepareEDVRESTProvider(opts *agentStartOpts, macCrypto *edv.MACCrypto) (*edv.RESTProvider, error) {
	userConf := opts.UserConfig
	capability := []byte(opts.EDVCapability)
	zcapSVC := zcapld.New(opts.AuthzKeyStoreURL, userConf.AccessToken, userConf.SecretShare)

	edvRESTProvider, err := edv.NewRESTProvider(opts.EDVServerURL, opts.EDVVaultID, macCrypto,
		edv.WithHeaders(func(req *http.Request) (*http.Header, error) {
			if len(capability) != 0 {
				action := "write"

				if req.Method == http.MethodGet {
					action = "read"
				}

				return zcapSVC.SignHeader(req, capability, action)
			}

			return nil, nil
		}),
		edv.WithFullDocumentsReturnedFromQueries())
	if err != nil {
		return nil, fmt.Errorf("failed to create new EDV REST provider: %w", err)
	}

	return edvRESTProvider, nil
}

func prepareFormattedProvider(opts *agentStartOpts, kmsStorageProvider storage.Provider, kmsImpl kms.KeyManager,
	cryptoImpl cryptoapi.Crypto, macCrypto *edv.MACCrypto,
	provider storage.Provider) (*formattedstore.FormattedProvider, error) {
	jweEncrypter, err := prepareJWEEncrypter(opts, kmsStorageProvider, kmsImpl, cryptoImpl)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare JWE encrypter: %w", err)
	}

	jweDecrypter := jose.NewJWEDecrypt(nil, cryptoImpl, kmsImpl)

	encryptedFormatter := edv.NewEncryptedFormatter(jweEncrypter, jweDecrypter, macCrypto)

	var o []formattedstore.Option

	if opts.UseEDVCache {
		clearCache := opts.EDVClearCache
		if clearCache == "" {
			clearCache = defaultClearCache
		}

		t, err := time.ParseDuration(clearCache)
		if err != nil {
			return nil, err
		}

		p, err := jsindexeddbcache.NewProvider("cache", t)
		if err != nil {
			return nil, err
		}

		o = append(o, formattedstore.WithCacheProvider(p))
	}

	if opts.UseEDVBatch {
		o = append(o, formattedstore.WithBatchWrite(opts.EDVBatchSize))
	}

	return formattedstore.NewFormattedProvider(provider, encryptedFormatter, false, o...), nil
}

func prepareMACCrypto(opts *agentStartOpts, kmsStorageProvider storage.Provider, kmsImpl kms.KeyManager,
	cryptoImpl cryptoapi.Crypto) (*edv.MACCrypto, error) {
	var (
		macKeyHandle interface{}
		err          error
	)

	if opts.KMSType == kmsTypeWebKMS {
		macKeyHandle = prepareRemoteKeyURL(opts, kms.HMACSHA256Tag256Type)
	} else {
		_, macKeyHandle, err = prepareLocalKeyHandle(kmsStorageProvider, kmsImpl, hmacKeyIDDBKeyName,
			kms.HMACSHA256Tag256Type)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare local MAC key handle: %w", err)
		}
	}

	return edv.NewMACCrypto(macKeyHandle, cryptoImpl), nil
}

func prepareJWEEncrypter(opts *agentStartOpts, kmsStorageProvider storage.Provider, kmsImpl kms.KeyManager,
	crypto cryptoapi.Crypto) (*jose.JWEEncrypt, error) {
	var (
		pubKeyBytes  []byte
		jweCryptoKID string
		err          error
	)

	if opts.KMSType == kmsTypeWebKMS {
		pubKeyBytes, jweCryptoKID, err = prepareRemoteJWEKey(opts.EDVOpsKIDURL, kmsImpl)
		if err != nil {
			return nil, err
		}
	} else {
		pubKeyBytes, jweCryptoKID, err = prepareLocalJWEKey(kmsStorageProvider, kmsImpl)
		if err != nil {
			return nil, err
		}
	}

	ecPubKey := new(cryptoapi.PublicKey)

	ecPubKey.KID = jweCryptoKID

	err = json.Unmarshal(pubKeyBytes, ecPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWE public key bytes to an EC public key object: %w", err)
	}

	jweEncrypter, err := jose.NewJWEEncrypt(jose.A256GCM, jose.DIDCommEncType, "", nil,
		[]*cryptoapi.PublicKey{ecPubKey}, crypto)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWE encrypter: %w", err)
	}

	return jweEncrypter, nil
}

func prepareRemoteJWEKey(keyURL string, kmsImpl kms.KeyManager) ([]byte, string, error) {
	id := strings.LastIndex(keyURL, "/keys/") + len("/keys/")
	if id > len(keyURL) {
		return nil, "", fmt.Errorf("prepreRemoteJWEKey: keyURL not well well formatted: %s", keyURL)
	}

	kid := keyURL[id:] // need KID part only of keyURL since remoteKMS has the keystore URL.

	pubKeyBytes, err := kmsImpl.ExportPubKeyBytes(kid)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve main public key bytes from remote KMS: %w", err)
	}

	return pubKeyBytes, kid, nil
}

func prepareLocalJWEKey(kmsStorageProvider storage.Provider, kmsImpl kms.KeyManager) ([]byte, string, error) {
	jweCryptoKID, jweCryptoKeyHandle, err := prepareLocalKeyHandle(kmsStorageProvider, kmsImpl, ecdhesKeyIDDBKeyName,
		kms.ECDH256KWAES256GCMType)
	if err != nil {
		return nil, "", fmt.Errorf("failed to prepare key handle for JWE crypto operations: %w", err)
	}

	pubKH, err := jweCryptoKeyHandle.Public()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get public key handle from JWE crypto private key handle: %w", err)
	}

	buf := new(bytes.Buffer)
	pubKeyWriter := keyio.NewWriter(buf)

	err = pubKH.WriteWithNoSecrets(pubKeyWriter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to write JWE public key to bytes: %w", err)
	}

	return buf.Bytes(), jweCryptoKID, nil
}

func prepareRemoteKeyURL(opts *agentStartOpts, keyType kms.KeyType) string {
	switch keyType { // nolint:exhaustive // no need to check for other key types, only HMAC key is a special case.
	case kms.HMACSHA256Tag256Type:
		return opts.EDVHMACKIDURL
	case kms.ECDH256KWAES256GCMType:
		return opts.EDVOpsKIDURL
	default:
		return opts.EDVOpsKIDURL
	}
}

func prepareLocalKeyHandle(storeProvider storage.Provider, keyManager kms.KeyManager,
	keyIDDBKeyName string, keyType kms.KeyType) (string, *keyset.Handle, error) {
	keyIDStore, err := storeProvider.OpenStore(keyIDStoreName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open key ID store: %w", err)
	}

	keyIDBytes, err := keyIDStore.Get(keyIDDBKeyName)
	if errors.Is(err, storage.ErrDataNotFound) {
		logger.Infof("No key handle ID was found in store %s with store DB key ID %s.", keyIDStoreName, keyIDDBKeyName)

		keyID, keyHandleUntyped, createErr := keyManager.Create(keyType)
		if createErr != nil {
			return "", nil, fmt.Errorf("failed to create new key: %w", createErr)
		}

		kh, ok := keyHandleUntyped.(*keyset.Handle)
		if !ok {
			return "", nil, errors.New("unable to assert newly created key handle as a key set handle pointer")
		}

		err = keyIDStore.Put(keyIDDBKeyName, []byte(keyID))
		if err != nil {
			return "", nil, fmt.Errorf("failed to put newly created key ID into key ID store: %w", err)
		}

		logger.Infof("Created new key handle and stored the key handle ID %s in store %s with store DB key ID %s.",
			keyID, keyIDStoreName, keyIDDBKeyName)

		return keyID, kh, nil
	} else if err != nil {
		return "", nil, fmt.Errorf("failed to key key ID bytes from key ID store: %w", err)
	}

	logger.Infof("Found existing key handle ID under store %s with store DB key ID %s.", keyIDStoreName, keyIDDBKeyName)

	keyID := string(keyIDBytes)

	keyHandleUntyped, getErr := keyManager.Get(keyID)
	if getErr != nil {
		return "", nil, fmt.Errorf("failed to get key handle from key manager: %w", getErr)
	}

	kh, ok := keyHandleUntyped.(*keyset.Handle)
	if !ok {
		return "", nil, errors.New("unable to assert key handle as a key set handle pointer")
	}

	return keyID, kh, nil
}

func setLogLevel(logLevel string) error {
	if logLevel != "" {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			return err
		}

		ariesLoglevel, err := arieslog.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("parse aries log level '%s' : %w", logLevel, err)
		}

		log.SetLevel("", level)
		arieslog.SetLevel("", ariesLoglevel)

		logger.Infof("log level set to `%s`", logLevel)
	}

	return nil
}

// jsNotifier notifies about all incoming events.
type jsNotifier struct {
}

// Notify is mock implementation of webhook notifier Notify().
func (n *jsNotifier) Notify(topic string, message []byte) error {
	payload := make(map[string]interface{})
	if err := json.Unmarshal(message, &payload); err != nil {
		return err
	}

	out, err := json.Marshal(&result{
		ID:      uuid.New().String(),
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		return err
	}

	js.Global().Call(handleResultFn, string(out))

	return nil
}

func postInitMsg() {
	if isTest {
		return
	}

	out, err := json.Marshal(&result{
		ID:    uuid.New().String(),
		Topic: wasmStartupTopic,
	})
	if err != nil {
		panic(err)
	}

	js.Global().Call(handleResultFn, string(out))
}
