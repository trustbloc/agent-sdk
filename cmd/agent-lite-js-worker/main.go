//go:build js && wasm
// +build js,wasm

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"compress/gzip"
	goctx "context"
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

	"github.com/btcsuite/btcd/btcec"
	"github.com/google/tink/go/subtle/random"
	"github.com/google/uuid"

	"github.com/hyperledger/aries-framework-go-ext/component/vdr/orb"
	"github.com/hyperledger/aries-framework-go/component/storage/indexeddb"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	kmscmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/kms"
	vcwalletcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/vcwallet"
	vdrcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/controller/command/vcwallet"
	cryptoapi "github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	webcrypto "github.com/hyperledger/aries-framework-go/pkg/crypto/webkms"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ldcontext/remote"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	vdrapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/webkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	ldstore "github.com/hyperledger/aries-framework-go/pkg/store/ld"
	"github.com/hyperledger/aries-framework-go/pkg/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/httpbinding"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/peer"
	"github.com/hyperledger/aries-framework-go/spi/storage"
	"github.com/mitchellh/mapstructure"

	jsonld "github.com/piprate/json-gold/ld"
	"github.com/trustbloc/auth/spi/gnap/proof/httpsig"
	"github.com/trustbloc/edge-core/pkg/log"
	edvclient "github.com/trustbloc/edv/pkg/client"
	"github.com/trustbloc/edv/pkg/restapi/models"

	"github.com/trustbloc/agent-sdk/pkg/auth/zcapld"
)

var logger = log.New("agent-js-worker")

const (
	wasmStartupTopic         = "asset-ready"
	handleResultFn           = "handleResult"
	commandPkg               = "agent"
	startFn                  = "Start"
	stopFn                   = "Stop"
	workers                  = 2
	kmsTypeWebKMS            = "webkms"
	masterKeyStoreName       = "MasterKey"
	masterKeyDBKeyName       = masterKeyStoreName
	masterKeyNumBytes        = 32
	walletTokenExpiryMins    = "20"
	authBootstrapDataPath    = "/gnap/bootstrap"
	defaultEndpoint          = "didcomm:transport/queue"
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
// nolint:lll
type agentStartOpts struct {
	HTTPResolvers            []string    `json:"http-resolver-url"`
	LogLevel                 string      `json:"log-level"`
	IndexedDBNamespace       string      `json:"indexedDB-namespace"`
	EDVServerURL             string      `json:"edvServerURL"`            // TODO to be removed/refined after universal wallet migration
	EDVVaultID               string      `json:"edvVaultID"`              // TODO to be removed/refined after universal wallet migration
	EDVCapability            string      `json:"edvCapability,omitempty"` // TODO to be removed/refined after universal wallet migration
	BlocDomain               string      `json:"blocDomain"`
	AuthzKeyStoreURL         string      `json:"authzKeyStoreURL,omitempty"`
	OpsKeyStoreURL           string      `json:"opsKeyStoreURL,omitempty"` // TODO to be removed/refined after universal wallet migration
	EDVOpsKIDURL             string      `json:"edvOpsKIDURL,omitempty"`   // TODO to be removed/refined after universal wallet migration
	EDVHMACKIDURL            string      `json:"edvHMACKIDURL,omitempty"`  // TODO to be removed/refined after universal wallet migration
	KMSType                  string      `json:"kmsType"`                  // TODO to be removed/refined after universal wallet migration
	UserConfig               *userConfig `json:"userConfig,omitempty"`
	EDVClearCache            string      `json:"edvClearCache"`
	EDVBatchSize             int         `json:"edvBatchSize"`
	CacheSize                int         `json:"cacheSize"`
	OPSKMSCapability         string      `json:"opsKMSCapability,omitempty"` // TODO to be removed/refined after universal wallet migration
	ContextProviderURLs      []string    `json:"context-provider-url"`
	UnanchoredDIDMaxLifeTime int         `json:"unanchoredDIDMaxLifeTime"`
	MediaTypeProfiles        []string    `json:"media-type-profiles"`
	AuthServerURL            string      `json:"hubAuthURL"`
	KMSServerURL             string      `json:"kms-server-url"`
	GNAPSigningJWK           string      `json:"gnap-signing-jwk"`
	GNAPAccessToken          string      `json:"gnap-access-token"`
	GNAPUserSubject          string      `json:"gnap-user-subject"`
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

	// Upon the first call `btcec.S256()` deserializes the pre-computed byte points for the secp256k1 curve and
	// it takes some time. Triggering that function here speeds up the following protocols.
	go initS256()

	for w := 0; w < workers; w++ {
		go worker(input, output, handlers)
	}
}

func initS256() {
	btcec.S256()
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
		opts, err := startOpts(c.Payload)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		err = setLogLevel(opts.LogLevel)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		ctx, err := agentOpts(opts)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		handlers, err := getAriesHandlers(ctx, opts)
		if err != nil {
			return newErrResult(c.ID, err.Error())
		}

		// add command handlers
		addCommandHandlers(handlers, pkgMap)

		// add stop agent handler
		//addStopAgentHandler(a, pkgMap)

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

func getAriesHandlers(ctx *walletLiteProvider,
	opts *agentStartOpts) ([]commandHandler, error) {

	wallet := vcwallet.New(ctx, &vcwalletcmd.Config{
		WebKMSCacheSize:                  opts.CacheSize,
		EDVReturnFullDocumentsOnQuery:    true,
		EDVBatchEndpointExtensionEnabled: true,
		WebKMSAuthzProvider:              &webkmsZCAPSigner{},
		EdvAuthzProvider:                 &edvZCAPSigner{},
	})

	vcmd, err := vdrcmd.New(ctx)
	if err != nil {
		return nil, err
	}

	kcmd := kmscmd.New(ctx)

	handlers := wallet.GetHandlers()
	handlers = append(handlers, vcmd.GetHandlers()...)
	handlers = append(handlers, kcmd.GetHandlers()...)

	var hh []commandHandler

	for _, h := range handlers {
		handle := h.Handle()

		hh = append(hh, commandHandler{
			name:   h.Name(),
			method: h.Method(),
			exec: func(rw io.Writer, req io.Reader) error {
				e := handle(rw, req)
				if e != nil {
					return fmt.Errorf("code: %+v, message: %w", e.Code(), e)
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
	logger.Debugf("agent start options: %+v\n", payload)

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

func createVDRs(resolvers []string, trustblocDomain string, unanchoredDIDMaxLifeTime int) ([]vdrapi.VDR, error) {
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

	VDRs := make([]vdrapi.VDR, len(set), len(set)+1)

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

	orbOpts := make([]orb.Option, 0)
	if unanchoredDIDMaxLifeTime > 0 {
		orbOpts = append(orbOpts, orb.WithUnanchoredMaxLifeTime(time.Duration(unanchoredDIDMaxLifeTime)*time.Second))
	}

	orbOpts = append(orbOpts, orb.WithDomain(trustblocDomain), orb.WithHTTPClient(http.DefaultClient))

	blocVDR, err := orb.New(nil, orbOpts...)
	if err != nil {
		return nil, err
	}

	VDRs = append(VDRs, blocVDR, key.New())

	return VDRs, nil
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

//nolint:gocyclo,funlen
func agentOpts(startOpts *agentStartOpts) (*walletLiteProvider, error) {
	var options []aries.Option
	provider := &walletLiteProvider{}

	if len(startOpts.MediaTypeProfiles) > 0 {
		provider.mediaTypeProfiles = startOpts.MediaTypeProfiles
	}

	// indexedDBProvider used by localKMS and JSON-LD contexts
	indexedDBProvider, err := createIndexedDBStorage(startOpts)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while creating IndexDB storage provider: %w", err)
	}

	provider.storageProvider = indexedDBProvider

	loader, err := createJSONLDDocumentLoader(indexedDBProvider, startOpts.ContextProviderURLs)
	if err != nil {
		return nil, fmt.Errorf("create document loader: %w", err)
	}

	provider.jSONLDDocumentLoader = loader

	var (
		cryptoImpl cryptoapi.Crypto
	)

	kmsImpl, cryptoImpl, options, err := createKMSAndCrypto(startOpts, indexedDBProvider, options)
	if err != nil {
		return nil, fmt.Errorf("unexpected failure while creating LocalKMS and Crypto instance: %w", err)
	}

	provider.kms = kmsImpl
	provider.crypto = cryptoImpl

	vdrs, err := createVDRs(startOpts.HTTPResolvers, startOpts.BlocDomain, startOpts.UnanchoredDIDMaxLifeTime)
	if err != nil {
		return nil, err
	}

	vrd, err := createVDR(vdrs, startOpts, provider.storageProvider)

	provider.vDRegistry = vrd

	return provider, nil
}

func createVDR(vdrs []vdrapi.VDR, startOpts *agentStartOpts, storageProvider storage.Provider) (*vdr.Registry, error) {
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

func createIndexedDBStorage(opts *agentStartOpts) (*indexeddb.Provider, error) {
	indexedDBKMSProvider, err := indexeddb.NewProvider(opts.IndexedDBNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create IndexedDB provider: %w", err)
	}

	return indexedDBKMSProvider, nil
}

func createKMSAndCrypto(opts *agentStartOpts, indexedDBKMSProvider storage.Provider,
	allAriesOptions []aries.Option) (kms.KeyManager, cryptoapi.Crypto, []aries.Option, error) {
	if opts.KMSType == kmsTypeWebKMS {
		return createWebKMS(opts, allAriesOptions)
	}

	return createLocalKMS(indexedDBKMSProvider, allAriesOptions)
}

//nolint:funlen,nestif
func createWebKMS(opts *agentStartOpts,
	allAriesOptions []aries.Option) (*webkms.RemoteKMS, *webcrypto.RemoteCrypto, []aries.Option, error) {
	var (
		headerFunc func(req *http.Request) (*http.Header, error)
		webKMS     *webkms.RemoteKMS
	)

	httpClient := http.DefaultClient

	if opts.GNAPAccessToken != "" {
		var err error

		if headerFunc, err = gnapAddHeaderFunc(opts.GNAPAccessToken, opts.GNAPSigningJWK); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create gnap header func: %w", err)
		}

		if opts.OpsKeyStoreURL == "" { // user requires onboarding
			keyStoreURL, _, err := webkms.CreateKeyStore(httpClient, opts.KMSServerURL, opts.GNAPUserSubject, "", nil,
				webkms.WithHeaders(headerFunc))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create key store: %w", err)
			}

			logger.Debugf("key store URL: %s", keyStoreURL)

			opts.OpsKeyStoreURL = keyStoreURL

			webKMS = webkms.New(keyStoreURL, httpClient, webkms.WithHeaders(headerFunc))

			if err := onboardUser(opts, webKMS); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to onboard user: %w", err)
			}
		} else {
			webKMS = webkms.New(opts.OpsKeyStoreURL, httpClient, webkms.WithHeaders(headerFunc))
		}
	} else if opts.OPSKMSCapability != "" {
		capability, err := decodeAndGunzip(opts.OPSKMSCapability)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to prepare OPS KMS capability for use: %w", err)
		}

		zcapSvc := zcapld.New(opts.AuthzKeyStoreURL, opts.UserConfig.AccessToken, opts.UserConfig.SecretShare)

		headerFunc = func(req *http.Request) (*http.Header, error) {
			invocationAction, err := capabilityInvocationAction(req)
			if err != nil {
				return nil, fmt.Errorf("webkms: failed to determine the capability's invocation action: %w", err)
			}

			return zcapSvc.SignHeader(req, capability, invocationAction)
		}

		webKMS = webkms.New(opts.OpsKeyStoreURL, httpClient, webkms.WithHeaders(headerFunc))
	}

	allAriesOptions = append(allAriesOptions,
		aries.WithKMS(func(ctx kms.Provider) (kms.KeyManager, error) {
			return webKMS, nil
		}))

	var kmsOpts []webkms.Opt

	if opts.CacheSize >= 0 {
		kmsOpts = append(kmsOpts, webkms.WithCache(opts.CacheSize))
	}

	kmsOpts = append(kmsOpts, webkms.WithHeaders(headerFunc))

	wCrypto := webcrypto.New(opts.OpsKeyStoreURL, http.DefaultClient, kmsOpts...)

	allAriesOptions = append(allAriesOptions, aries.WithCrypto(wCrypto))

	return webKMS, wCrypto, allAriesOptions, nil
}

func gnapAddHeaderFunc(gnapAccessToken, gnapSigningJWK string) (func(req *http.Request) (*http.Header, error), error) {
	signingJWK := &jwk.JWK{}

	err := json.Unmarshal([]byte(gnapSigningJWK), signingJWK)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal gnap signing jwk: %w", err)
	}

	signingJWK.Algorithm = "ES256"

	return func(req *http.Request) (*http.Header, error) {
		req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", gnapAccessToken))

		r, err := httpsig.Sign(req, nil, signingJWK, "sha-256") // TODO: pass body bytes
		if err != nil {
			return nil, fmt.Errorf("failed to sign request: %w", err)
		}

		return &r.Header, nil
	}, nil
}

type bootstrapData struct {
	User              string `json:"user,omitempty"`
	UserEDVVaultURL   string `json:"edvVaultURL,omitempty"`
	OpsEDVVaultURL    string `json:"opsVaultURL,omitempty"`
	AuthzKeyStoreURL  string `json:"authzKeyStoreURL,omitempty"`
	OpsKeyStoreURL    string `json:"opsKeyStoreURL,omitempty"`
	EDVOpsKIDURL      string `json:"edvOpsKIDURL,omitempty"`
	EDVHMACKIDURL     string `json:"edvHMACKIDURL,omitempty"`
	UserEDVCapability string `json:"edvCapability,omitempty"`
	OPSKMSCapability  string `json:"opsKMSCapability,omitempty"`
	UserEDVServer     string `json:"userEDVServer,omitempty"`
	UserEDVVaultID    string `json:"userEDVVaultID,omitempty"`
	UserEDVEncKID     string `json:"userEDVEncKID,omitempty"`
	UserEDVMACKID     string `json:"userEDVMACKID,omitempty"`
	TokenExpiry       string `json:"tokenExpiry,omitempty"`
}

type userBootstrapData struct {
	Data *bootstrapData `json:"data,omitempty"`
}

//nolint:funlen
func onboardUser(opts *agentStartOpts, webKMS *webkms.RemoteKMS) error {
	// 1. Create EDV data vault
	config := &models.DataVaultConfiguration{
		Sequence:    0,
		Controller:  opts.GNAPUserSubject,
		ReferenceID: uuid.New().String(),
		KEK:         models.IDTypePair{ID: uuid.New().URN(), Type: "AesKeyWrappingKey2019"},
		HMAC:        models.IDTypePair{ID: uuid.New().URN(), Type: "Sha256HmacKey2019"},
	}

	edvVaultURL, _, err := edvclient.New(opts.EDVServerURL).CreateDataVault(config,
		edvclient.WithRequestHeader(func(req *http.Request) (*http.Header, error) {
			req.Header.Set("Authorization", "GNAP "+opts.GNAPAccessToken)

			return &req.Header, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create EDV data vault: %w", err)
	}

	opts.EDVVaultID = getVaultID(edvVaultURL)

	logger.Debugf("EDV data vault URL: %s", edvVaultURL)

	// 2. Create key for encrypted formatter (EDV storage provider)
	edvOpsKID, edvOpsKeyURL, err := webKMS.Create(kms.NISTP256ECDHKW)
	if err != nil {
		return fmt.Errorf("failed to create EDV ops key: %w", err)
	}

	opts.EDVOpsKIDURL = edvOpsKeyURL.(string) //nolint:errcheck,forcetypeassert

	logger.Debugf("EDV ops key URL: %s", opts.EDVOpsKIDURL)

	// 3. Create MAC key for generating document IDs (EDV storage provider)
	edvHMACKID, edvHMACKeyURL, err := webKMS.Create(kms.HMACSHA256Tag256)
	if err != nil {
		return fmt.Errorf("failed to create EDV HMAC key: %w", err)
	}

	opts.EDVHMACKIDURL = edvHMACKeyURL.(string) //nolint:errcheck,forcetypeassert

	logger.Debugf("EDV HMAC key URL: %s", opts.EDVHMACKIDURL)

	// 4. Post bootstrap data to auth server
	data := &bootstrapData{
		User:              uuid.NewString(),
		UserEDVVaultURL:   edvVaultURL,
		OpsEDVVaultURL:    "",
		AuthzKeyStoreURL:  "",
		OpsKeyStoreURL:    opts.OpsKeyStoreURL,
		EDVOpsKIDURL:      opts.EDVOpsKIDURL,
		EDVHMACKIDURL:     opts.EDVHMACKIDURL,
		UserEDVCapability: "",
		OPSKMSCapability:  "",
		UserEDVVaultID:    opts.EDVVaultID,
		UserEDVServer:     opts.EDVServerURL,
		UserEDVEncKID:     edvOpsKID,
		UserEDVMACKID:     edvHMACKID,
		TokenExpiry:       walletTokenExpiryMins,
	}

	reqBytes, err := json.Marshal(userBootstrapData{
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("marshal boostrap data : %w", err)
	}

	req, err := http.NewRequestWithContext(goctx.Background(),
		http.MethodPost, opts.AuthServerURL+authBootstrapDataPath, bytes.NewBuffer(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to create bootstrap request to auth server: %w", err)
	}

	req.Header.Set("Authorization", "GNAP "+opts.GNAPAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post bootstrap data: %w", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body: %w", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to post bootstrap data: %s", resp.Status)
	}

	return nil
}

func getVaultID(vaultURL string) string {
	parts := strings.Split(vaultURL, "/")

	return parts[len(parts)-1]
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

func capabilityInvocationAction(req *http.Request) (string, error) { //nolint:funlen,gocognit,gocyclo
	s := strings.Split(req.URL.Path, "/")

	const minPathLen = 5 // /v1/keystores/{key_store_id}/keys

	if len(s) < minPathLen {
		return "", errors.New("invalid path")
	}

	op := strings.ToLower(s[4])

	var action string

	switch op {
	case "keys":
		op = strings.ToLower(s[len(s)-1])

		switch op {
		case "sign":
			if req.Method == http.MethodPost {
				action = "sign"
			}
		case "verify":
			if req.Method == http.MethodPost {
				action = "verify"
			}
		case "encrypt":
			if req.Method == http.MethodPost {
				action = "encrypt"
			}
		case "decrypt":
			if req.Method == http.MethodPost {
				action = "decrypt"
			}
		case "computemac":
			if req.Method == http.MethodPost {
				action = "computeMAC"
			}
		case "verifymac":
			if req.Method == http.MethodPost {
				action = "verifyMAC"
			}
		case "signmulti":
			if req.Method == http.MethodPost {
				action = "signMulti"
			}
		case "verifymulti":
			if req.Method == http.MethodPost {
				action = "verifyMulti"
			}
		case "deriveproof":
			if req.Method == http.MethodPost {
				action = "deriveProof"
			}
		case "verifyproof":
			if req.Method == http.MethodPost {
				action = "verifyProof"
			}
		case "easy":
			if req.Method == http.MethodPost {
				action = "easy"
			}
		case "wrap": //nolint:goconst
			if req.Method == http.MethodPost {
				action = "wrap"
			}
		case "unwrap":
			if req.Method == http.MethodPost {
				action = "unwrap"
			}
		default:
			if req.Method == http.MethodPost {
				action = "createKey"
			}

			if req.Method == http.MethodPut {
				action = "importKey"
			}

			if req.Method == http.MethodGet && op != "keys" {
				action = "exportKey"
			}
		}
	case "wrap":
		if req.Method == http.MethodPost {
			action = "wrap"
		}
	case "easyopen":
		if req.Method == http.MethodPost {
			action = "easyOpen"
		}
	case "sealopen":
		if req.Method == http.MethodPost {
			action = "sealOpen"
		}
	}

	if action == "" {
		return "", fmt.Errorf("unsupported operation: %s /%s", req.Method, op)
	}

	return action, nil
}

func createLocalKMS(indexedDBKMSProvider storage.Provider,
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

type webkmsZCAPSigner struct{}

func (b *webkmsZCAPSigner) GetHeaderSigner(authzKeyStoreURL, accessToken, secretShare string) vcwalletcmd.HTTPHeaderSigner { //nolint:lll
	return &webKMSHTTPHeaderSigner{
		zcapSVC: zcapld.New(authzKeyStoreURL, accessToken, secretShare),
	}
}

type edvZCAPSigner struct{}

func (b *edvZCAPSigner) GetHeaderSigner(authzKeyStoreURL, accessToken, secretShare string) vcwalletcmd.HTTPHeaderSigner { //nolint:lll
	return &edvHTTPHeaderSigner{
		zcapSVC: zcapld.New(authzKeyStoreURL, accessToken, secretShare),
	}
}

// webKMSHTTPHeaderSigner is zcap based http header signer for vc wallet webkms header.
type webKMSHTTPHeaderSigner struct {
	zcapSVC *zcapld.Service
}

// SignHeader signs HTTP header based on zcap.
func (w *webKMSHTTPHeaderSigner) SignHeader(req *http.Request, kmsCapability []byte) (*http.Header, error) {
	capability, err := decodeAndGunzip(string(kmsCapability))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare KMS capability for use: %w ", err)
	}

	if len(capability) != 0 {
		invocationAction, err := capabilityInvocationAction(req)
		if err != nil {
			return nil, fmt.Errorf("webkms: failed to determine the capability's invocation action: %w", err)
		}

		return w.zcapSVC.SignHeader(req, capability, invocationAction)
	}

	return &req.Header, nil
}

// edvHTTPHeaderSigner is zcap based http header signer for vc wallet edv header.
type edvHTTPHeaderSigner struct {
	zcapSVC *zcapld.Service
}

// SignHeader signs HTTP header based on zcap.
func (w *edvHTTPHeaderSigner) SignHeader(req *http.Request, capability []byte) (*http.Header, error) {
	if len(capability) != 0 {
		action := "write"

		if req.Method == http.MethodGet {
			action = "read"
		}

		return w.zcapSVC.SignHeader(req, capability, action)
	}

	return &req.Header, nil
}

type ldStoreProvider struct {
	ContextStore        ldstore.ContextStore
	RemoteProviderStore ldstore.RemoteProviderStore
}

func (p *ldStoreProvider) JSONLDContextStore() ldstore.ContextStore {
	return p.ContextStore
}

func (p *ldStoreProvider) JSONLDRemoteProviderStore() ldstore.RemoteProviderStore {
	return p.RemoteProviderStore
}

func createJSONLDDocumentLoader(storageProvider storage.Provider,
	contextProviderURLs []string) (jsonld.DocumentLoader, error) {
	contextStore, err := ldstore.NewContextStore(storageProvider)
	if err != nil {
		return nil, fmt.Errorf("create JSON-LD context store: %w", err)
	}

	remoteProviderStore, err := ldstore.NewRemoteProviderStore(storageProvider)
	if err != nil {
		return nil, fmt.Errorf("create remote provider store: %w", err)
	}

	ldStore := &ldStoreProvider{
		ContextStore:        contextStore,
		RemoteProviderStore: remoteProviderStore,
	}

	var loaderOpts []ld.DocumentLoaderOpts

	if len(contextProviderURLs) > 0 {
		for _, url := range contextProviderURLs {
			loaderOpts = append(loaderOpts, ld.WithRemoteProvider(remote.NewProvider(url)))
		}
	} else {
		// fetching contexts from the network is enabled if no context providers are specified
		loaderOpts = append(loaderOpts,
			ld.WithRemoteDocumentLoader(jsonld.NewDefaultDocumentLoader(http.DefaultClient)))
	}

	documentLoader, err := ld.NewDocumentLoader(ldStore, loaderOpts...)
	if err != nil {
		return nil, fmt.Errorf("new document loader: %w", err)
	}

	return documentLoader, nil
}
