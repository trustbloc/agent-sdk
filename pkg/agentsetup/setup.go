//go:build js && wasm
// +build js,wasm

package agentsetup

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import (
	"bytes"
	goctx "context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/subtle/random"
	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/orb"
	"github.com/hyperledger/aries-framework-go/component/storage/edv"
	"github.com/hyperledger/aries-framework-go/component/storage/indexeddb"
	"github.com/hyperledger/aries-framework-go/component/storageutil/batchedstore"
	"github.com/hyperledger/aries-framework-go/component/storageutil/cachedstore"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	cryptoapi "github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto/primitive/composite/keyio"
	webcrypto "github.com/hyperledger/aries-framework-go/pkg/crypto/webkms"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/packer"
	arieshttp "github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/http"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ldcontext/remote"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/webkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	ldstore "github.com/hyperledger/aries-framework-go/pkg/store/ld"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/httpbinding"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/web"
	"github.com/hyperledger/aries-framework-go/spi/storage"
	jsonld "github.com/piprate/json-gold/ld"
	"github.com/trustbloc/auth/spi/gnap/proof/httpsig"
	"github.com/trustbloc/edge-core/pkg/log"
	edvclient "github.com/trustbloc/edv/pkg/client"
	"github.com/trustbloc/edv/pkg/restapi/models"

	"github.com/trustbloc/agent-sdk/pkg/storage/jsindexeddbcache"
)

var logger = log.New("agent-js-worker")

const (
	StorageTypeIndexedDB     = "indexedDB"
	StorageTypeEDV           = "edv"
	validStorageTypesMsg     = "Valid storage types: " + StorageTypeEDV + ", " + StorageTypeIndexedDB
	blankStorageTypeErrMsg   = "no storage type specified. " + validStorageTypesMsg
	InvalidStorageTypeErrMsg = "%s is not a valid storage type. " + validStorageTypesMsg
	kmsTypeWebKMS            = "webkms"
	hmacKeyIDDBKeyName       = "hmackeyid"
	keyIDStoreName           = "keyid"
	ecdhesKeyIDDBKeyName     = "ecdheskeyid"
	masterKeyStoreName       = "MasterKey"
	masterKeyDBKeyName       = masterKeyStoreName
	masterKeyNumBytes        = 32
	defaultClearCache        = "5m"
	walletTokenExpiryMins    = "20"
	authBootstrapDataPath    = "/gnap/bootstrap"
)

// AgentStartOpts contains opts for starting agent.
// nolint:lll
type AgentStartOpts struct {
	Label                    string      `json:"agent-default-label"`
	HTTPResolvers            []string    `json:"http-resolver-url"`
	AutoAccept               bool        `json:"auto-accept"`
	OutboundTransport        []string    `json:"outbound-transport"`
	TransportReturnRoute     string      `json:"transport-return-route"`
	LogLevel                 string      `json:"log-level"`
	StorageType              string      `json:"storage-type"`
	IndexedDBNamespace       string      `json:"indexed-db-namespace"`
	EDVServerURL             string      `json:"edv-server-url"` // TODO to be removed/refined after universal wallet migration
	EDVVaultID               string      `json:"edv-vault-id"`   // TODO to be removed/refined after universal wallet migration
	BlocDomain               string      `json:"bloc-domain"`
	TrustblocResolver        string      `json:"trustbloc-resolver"`
	OpsKeyStoreURL           string      `json:"ops-key-store-url,omitempty"` // TODO to be removed/refined after universal wallet migration
	EDVOpsKIDURL             string      `json:"edv-ops-kid-url,omitempty"`   // TODO to be removed/refined after universal wallet migration
	EDVHMACKIDURL            string      `json:"edv-hmac-kid-url,omitempty"`  // TODO to be removed/refined after universal wallet migration
	KMSType                  string      `json:"kms-type"`                    // TODO to be removed/refined after universal wallet migration
	UserConfig               *UserConfig `json:"user-config,omitempty"`
	UseEDVCache              bool        `json:"use-edv-cache"`
	EDVClearCache            string      `json:"edv-clear-cache"`
	UseEDVBatch              bool        `json:"use-edv-batch"`
	EDVBatchSize             int         `json:"edv-batch-size"`
	CacheSize                int         `json:"cache-size"`
	DidAnchorOrigin          string      `json:"did-anchor-origin"`
	SidetreeToken            string      `json:"sidetree-token"`
	ContextProviderURLs      []string    `json:"context-provider-url"`
	UnanchoredDIDMaxLifeTime int         `json:"unanchored-din-max-life-time"`
	KeyType                  string      `json:"key-type"`
	KeyAgreementType         string      `json:"key-agreement-type"`
	MediaTypeProfiles        []string    `json:"media-type-profiles"`
	WebSocketReadLimit       int64       `json:"web-socket-read-limit"`
	AuthServerURL            string      `json:"hub-auth-url"`
	KMSServerURL             string      `json:"kms-server-url"`
	GNAPSigningJWK           string      `json:"gnap-signing-jwk"`
	GNAPAccessToken          string      `json:"gnap-access-token"`
	GNAPUserSubject          string      `json:"gnap-user-subject"`
}

type UserConfig struct {
	AccessToken string `json:"accessToken,omitempty"` // TODO should remove this
	SecretShare string `json:"walletSecretShare"`
}

type kmsProvider struct {
	store             kms.Store
	secretLockService secretlock.Service
}

func (k kmsProvider) StorageProvider() kms.Store {
	return k.store
}

func (k kmsProvider) SecretLock() secretlock.Service {
	return k.secretLockService
}

func CreateVDRs(resolvers []string, trustblocDomain string, unanchoredDIDMaxLifeTime int) ([]vdr.VDR, error) {
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

	orbOpts := make([]orb.Option, 0)
	if unanchoredDIDMaxLifeTime > 0 {
		orbOpts = append(orbOpts, orb.WithUnanchoredMaxLifeTime(time.Duration(unanchoredDIDMaxLifeTime)*time.Second))
	}

	orbOpts = append(orbOpts, orb.WithDomain(trustblocDomain), orb.WithHTTPClient(http.DefaultClient))

	blocVDR, err := orb.New(nil, orbOpts...)
	if err != nil {
		return nil, err
	}

	VDRs = append(VDRs, blocVDR, key.New(), web.New())

	return VDRs, nil
}

var (
	//nolint:gochecknoglobals // translation tables copied from afgo for consistency
	KeyTypes = map[string]kms.KeyType{
		"ed25519":           kms.ED25519Type,
		"ecdsap256ieee1363": kms.ECDSAP256TypeIEEEP1363,
		"ecdsap256der":      kms.ECDSAP256TypeDER,
		"ecdsap384ieee1363": kms.ECDSAP384TypeIEEEP1363,
		"ecdsap384der":      kms.ECDSAP384TypeDER,
		"ecdsap521ieee1363": kms.ECDSAP521TypeIEEEP1363,
		"ecdsap521der":      kms.ECDSAP521TypeDER,
	}

	//nolint:gochecknoglobals // translation tables copied from afgo for consistency
	KeyAgreementTypes = map[string]kms.KeyType{
		"x25519kw": kms.X25519ECDHKWType,
		"p256kw":   kms.NISTP256ECDHKWType,
		"p384kw":   kms.NISTP384ECDHKWType,
		"p521kw":   kms.NISTP521ECDHKWType,
	}
)

func AddOutboundTransports(startOpts *AgentStartOpts, options []aries.Option) ([]aries.Option, error) {
	for _, transport := range startOpts.OutboundTransport {
		switch transport {
		case "http":
			outbound, err := arieshttp.NewOutbound(arieshttp.WithOutboundHTTPClient(&http.Client{}))
			if err != nil {
				return nil, err
			}

			options = append(options, aries.WithOutboundTransports(outbound))
		case "ws":
			var outboundOpts []ws.OutboundClientOpt

			if startOpts.WebSocketReadLimit > 0 {
				outboundOpts = append(outboundOpts, ws.WithOutboundReadLimit(startOpts.WebSocketReadLimit))
			}

			options = append(options, aries.WithOutboundTransports(ws.NewOutbound(outboundOpts...)))
		default:
			return nil, fmt.Errorf("unsupported transport : %s", transport)
		}
	}

	return options, nil
}

func CreateIndexedDBStorage(opts *AgentStartOpts) (*indexeddb.Provider, error) {
	indexedDBKMSProvider, err := indexeddb.NewProvider(opts.IndexedDBNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create IndexedDB provider: %w", err)
	}

	return indexedDBKMSProvider, nil
}

func AddStorageOptions(startOpts *AgentStartOpts, indexedDBProvider *indexeddb.Provider,
	ariesKMS kms.KeyManager, ariesCrypto cryptoapi.Crypto, allAriesOptions []aries.Option) ([]aries.Option, error) {
	if startOpts.StorageType == "" {
		return nil, errors.New(blankStorageTypeErrMsg)
	}

	var store storage.Provider

	var err error

	switch startOpts.StorageType {
	case StorageTypeEDV:
		store, err = CreateEDVStorageProvider(startOpts, indexedDBProvider, ariesKMS, ariesCrypto)
		if err != nil {
			return nil, fmt.Errorf("failed to create storage: %w", err)
		}

		allAriesOptions = append(allAriesOptions, aries.WithProtocolStateStoreProvider(indexedDBProvider))
	case StorageTypeIndexedDB:
		store = indexedDBProvider
	default:
		return nil, fmt.Errorf(InvalidStorageTypeErrMsg, startOpts.StorageType)
	}

	allAriesOptions = append(allAriesOptions, aries.WithStoreProvider(store))

	return allAriesOptions, nil
}

func CreateKMSAndCrypto(opts *AgentStartOpts, indexedDBKMSProvider storage.Provider,
	allAriesOptions []aries.Option) (kms.KeyManager, cryptoapi.Crypto, []aries.Option, error) {
	if opts.KMSType == kmsTypeWebKMS {
		return createWebKMS(opts, allAriesOptions)
	}

	return createLocalKMS(indexedDBKMSProvider, allAriesOptions)
}

//nolint:funlen,nestif
func createWebKMS(opts *AgentStartOpts,
	allAriesOptions []aries.Option) (*webkms.RemoteKMS, *webcrypto.RemoteCrypto, []aries.Option, error) {
	var (
		headerFunc func(req *http.Request) (*http.Header, error)
		webKMS     *webkms.RemoteKMS
	)

	httpClient := http.DefaultClient

	if opts.GNAPAccessToken != "" {
		var err error

		if headerFunc, err = GNAPAddHeaderFunc(opts.GNAPAccessToken, opts.GNAPSigningJWK); err != nil {
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

func GNAPAddHeaderFunc(gnapAccessToken, gnapSigningJWK string) (func(req *http.Request) (*http.Header, error), error) {
	signingJWK := &jwk.JWK{}

	err := json.Unmarshal([]byte(gnapSigningJWK), signingJWK)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal gnap signing jwk: %w", err)
	}

	signingJWK.Algorithm = "ES256"

	return func(req *http.Request) (*http.Header, error) {
		req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", gnapAccessToken))

		var (
			body []byte
			e    error
		)

		if req.Body != nil {
			body, e = ioutil.ReadAll(req.Body)
			if e != nil {
				return nil, fmt.Errorf("failed to read body: %w", err)
			}

			req.Body = ioutil.NopCloser(bytes.NewReader(body))
		}

		r, err := httpsig.Sign(req, body, signingJWK, "sha-256")
		if err != nil {
			return nil, fmt.Errorf("failed to sign request: %w", err)
		}

		return &r.Header, nil
	}, nil
}

type bootstrapData struct {
	User            string `json:"user,omitempty"`
	UserEDVVaultURL string `json:"edvVaultURL,omitempty"`
	OpsEDVVaultURL  string `json:"opsVaultURL,omitempty"`
	OpsKeyStoreURL  string `json:"opsKeyStoreURL,omitempty"`
	EDVOpsKIDURL    string `json:"edvOpsKIDURL,omitempty"`
	EDVHMACKIDURL   string `json:"edvHMACKIDURL,omitempty"`
	UserEDVServer   string `json:"userEDVServer,omitempty"`
	UserEDVVaultID  string `json:"userEDVVaultID,omitempty"`
	UserEDVEncKID   string `json:"userEDVEncKID,omitempty"`
	UserEDVMACKID   string `json:"userEDVMACKID,omitempty"`
	TokenExpiry     string `json:"tokenExpiry,omitempty"`
}

type userBootstrapData struct {
	Data *bootstrapData `json:"data,omitempty"`
}

//nolint:funlen,gocyclo
func onboardUser(opts *AgentStartOpts, webKMS *webkms.RemoteKMS) error {
	// 1. Create EDV data vault
	config := &models.DataVaultConfiguration{
		Sequence:    0,
		Controller:  opts.GNAPUserSubject,
		ReferenceID: uuid.New().String(),
		KEK:         models.IDTypePair{ID: uuid.New().URN(), Type: "AesKeyWrappingKey2019"},
		HMAC:        models.IDTypePair{ID: uuid.New().URN(), Type: "Sha256HmacKey2019"},
	}

	var (
		headerFunc func(r2 *http.Request) (*http.Header, error)
		err        error
	)

	if opts.GNAPSigningJWK != "" && opts.GNAPAccessToken != "" {
		headerFunc, err = GNAPAddHeaderFunc(opts.GNAPAccessToken, opts.GNAPSigningJWK)
		if err != nil {
			return fmt.Errorf("failed to create gnap header func: %w", err)
		}
	} else {
		headerFunc = func(req *http.Request) (*http.Header, error) {
			req.Header.Set("Authorization", "GNAP "+opts.GNAPAccessToken)

			return &req.Header, nil
		}
	}

	edvVaultURL, _, err := edvclient.New(opts.EDVServerURL).CreateDataVault(config,
		edvclient.WithRequestHeader(headerFunc),
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
		User:            uuid.NewString(),
		UserEDVVaultURL: edvVaultURL,
		OpsEDVVaultURL:  "",
		OpsKeyStoreURL:  opts.OpsKeyStoreURL,
		EDVOpsKIDURL:    opts.EDVOpsKIDURL,
		EDVHMACKIDURL:   opts.EDVHMACKIDURL,
		UserEDVVaultID:  opts.EDVVaultID,
		UserEDVServer:   opts.EDVServerURL,
		UserEDVEncKID:   edvOpsKID,
		UserEDVMACKID:   edvHMACKID,
		TokenExpiry:     walletTokenExpiryMins,
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

	// TODO (#412): Create our own implementation of the KMS storage interface and pass it in here instead of wrapping
	//  the Aries storage provider.
	kmsStore, err := kms.NewAriesProviderWrapper(indexedDBKMSProvider)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create Aries KMS store wrapper")
	}

	kmsProv := kmsProvider{
		store:             kmsStore,
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

// CreateEDVStorageProvider creates an EDV storage provider. The given cryptoStorageProvider is used for storage of
// keys if a local KMS is used.
// The EDV provider returned by this function will used IndexedDB as a cache and will also make use of automatic
// batching to improve performance.
func CreateEDVStorageProvider(opts *AgentStartOpts, cryptoStorageProvider storage.Provider, kmsImpl kms.KeyManager,
	cryptoImpl cryptoapi.Crypto) (storage.Provider, error) {
	macCrypto, err := prepareMACCrypto(opts, cryptoStorageProvider, kmsImpl, cryptoImpl)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare MAC crypto: %w", err)
	}

	encryptedFormatter, err := prepareEncryptedFormatter(opts, cryptoStorageProvider, kmsImpl, cryptoImpl, macCrypto)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare formatted provider: %w", err)
	}

	edvRESTProvider, err := prepareEDVRESTProvider(opts, encryptedFormatter)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare edv rest provider: %w", err)
	}

	batchedProvider := batchedstore.NewProvider(edvRESTProvider, opts.EDVBatchSize)

	clearCache := opts.EDVClearCache
	if clearCache == "" {
		clearCache = defaultClearCache
	}

	t, err := time.ParseDuration(clearCache)
	if err != nil {
		return nil, err
	}

	indexedDBCacheProvider, err := jsindexeddbcache.NewProvider("cache", t)
	if err != nil {
		return nil, err
	}

	cachedProvider := cachedstore.NewProvider(batchedProvider, indexedDBCacheProvider)

	return cachedProvider, nil
}

//nolint:nestif
func prepareEDVRESTProvider(opts *AgentStartOpts, formatter *edv.EncryptedFormatter) (*edv.RESTProvider, error) {
	var headerFunc func(req *http.Request) (*http.Header, error)

	if opts.GNAPAccessToken != "" {
		var err error

		if headerFunc, err = GNAPAddHeaderFunc(opts.GNAPAccessToken, opts.GNAPSigningJWK); err != nil {
			return nil, fmt.Errorf("failed to create gnap header func: %w", err)
		}
	}

	return edv.NewRESTProvider(opts.EDVServerURL, opts.EDVVaultID, formatter,
		edv.WithHeaders(headerFunc),
		edv.WithFullDocumentsReturnedFromQueries(), edv.WithBatchEndpointExtension()), nil
}

func prepareEncryptedFormatter(opts *AgentStartOpts, kmsStorageProvider storage.Provider, kmsImpl kms.KeyManager,
	cryptoImpl cryptoapi.Crypto, macCrypto *edv.MACCrypto) (*edv.EncryptedFormatter, error) {
	jweEncrypter, err := prepareJWEEncrypter(opts, kmsStorageProvider, kmsImpl, cryptoImpl)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare JWE encrypter: %w", err)
	}

	jweDecrypter := jose.NewJWEDecrypt(nil, cryptoImpl, kmsImpl)

	return edv.NewEncryptedFormatter(jweEncrypter, jweDecrypter, macCrypto, edv.WithDeterministicDocumentIDs()), nil
}

func prepareMACCrypto(opts *AgentStartOpts, kmsStorageProvider storage.Provider, kmsImpl kms.KeyManager,
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

func prepareJWEEncrypter(opts *AgentStartOpts, kmsStorageProvider storage.Provider, kmsImpl kms.KeyManager,
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

	jweEncrypter, err := jose.NewJWEEncrypt(jose.A256GCM, packer.EnvelopeEncodingTypeV2, "", "", nil,
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

	pubKeyBytes, _, err := kmsImpl.ExportPubKeyBytes(kid)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve main public key bytes from remote KMS: %w", err)
	}

	return pubKeyBytes, kid, nil
}

func prepareLocalJWEKey(kmsStorageProvider storage.Provider, kmsImpl kms.KeyManager) ([]byte, string, error) {
	jweCryptoKID, jweCryptoKeyHandle, err := prepareLocalKeyHandle(kmsStorageProvider, kmsImpl, ecdhesKeyIDDBKeyName,
		kms.NISTP256ECDHKW)
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

func prepareRemoteKeyURL(opts *AgentStartOpts, keyType kms.KeyType) string {
	switch keyType { // nolint:exhaustive // no need to check for other key types, only HMAC key is a special case.
	case kms.HMACSHA256Tag256Type:
		return opts.EDVHMACKIDURL
	case kms.NISTP256ECDHKW:
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

func SetLogLevel(logLevel string) error {
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

func CreateJSONLDDocumentLoader(storageProvider storage.Provider,
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
