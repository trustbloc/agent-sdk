/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package startcmd  defines service commands and options.
package startcmd

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/mux"
	"github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb"
	"github.com/hyperledger/aries-framework-go-ext/component/storage/mongodb"
	"github.com/hyperledger/aries-framework-go-ext/component/storage/mysql"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/orb"
	"github.com/hyperledger/aries-framework-go/component/storage/leveldb"
	"github.com/hyperledger/aries-framework-go/component/storageutil/mem"
	"github.com/hyperledger/aries-framework-go/pkg/common/log"
	"github.com/hyperledger/aries-framework-go/pkg/controller"
	"github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport"
	arieshttp "github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/http"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/httpbinding"
	"github.com/hyperledger/aries-framework-go/spi/storage"
	"github.com/rs/cors"
	"github.com/spf13/cobra"

	sdkcontroller "github.com/trustbloc/agent-sdk/pkg/controller"
)

const (
	// api host flag.
	agentHostFlagName      = "api-host"
	agentHostEnvKey        = "ARIESD_API_HOST"
	agentHostFlagShorthand = "a"
	agentHostFlagUsage     = "Host Name:Port." +
		" Alternatively, this can be set with the following environment variable: " + agentHostEnvKey

	// api token flag.
	agentTokenFlagName      = "api-token"
	agentTokenEnvKey        = "ARIESD_API_TOKEN" // nolint:gosec
	agentTokenFlagShorthand = "t"
	agentTokenFlagUsage     = "Check for bearer token in the authorization header (optional)." +
		" Alternatively, this can be set with the following environment variable: " + agentTokenEnvKey

	databaseTypeFlagName      = "database-type"
	databaseTypeEnvKey        = "ARIESD_DATABASE_TYPE"
	databaseTypeFlagShorthand = "q"
	databaseTypeFlagUsage     = "The type of database to use for everything except key storage. " +
		"Supported options: mem, couchdb, mysql, leveldb, mongodb. " +
		" Alternatively, this can be set with the following environment variable: " + databaseTypeEnvKey

	databaseURLFlagName      = "database-url"
	databaseURLEnvKey        = "ARIESD_DATABASE_URL"
	databaseURLFlagShorthand = "v"
	databaseURLFlagUsage     = "The URL (or connection string) of the database. Not needed if using memstore." +
		" For CouchDB, include the username:password@ text if required. " +
		" Alternatively, this can be set with the following environment variable: " + databaseURLEnvKey

	databasePrefixFlagName      = "database-prefix"
	databasePrefixEnvKey        = "ARIESD_DATABASE_PREFIX"
	databasePrefixFlagShorthand = "u"
	databasePrefixFlagUsage     = "An optional prefix to be used when creating and retrieving underlying databases. " +
		" Alternatively, this can be set with the following environment variable: " + databasePrefixEnvKey

	databaseTimeoutFlagName  = "database-timeout"
	databaseTimeoutFlagUsage = "Total time in seconds to wait until the db is available before giving up." +
		" Default: " + databaseTimeoutDefault + " seconds." +
		" Alternatively, this can be set with the following environment variable: " + databaseTimeoutEnvKey
	databaseTimeoutEnvKey  = "ARIESD_DATABASE_TIMEOUT"
	databaseTimeoutDefault = "30"

	// webhook url flag.
	agentWebhookFlagName      = "webhook-url"
	agentWebhookEnvKey        = "ARIESD_WEBHOOK_URL"
	agentWebhookFlagShorthand = "w"
	agentWebhookFlagUsage     = "URL to send notifications to." +
		" This flag can be repeated, allowing for multiple listeners." +
		" Alternatively, this can be set with the following environment variable (in CSV format): " + agentWebhookEnvKey

	// default label flag.
	agentDefaultLabelFlagName      = "agent-default-label"
	agentDefaultLabelEnvKey        = "ARIESD_DEFAULT_LABEL"
	agentDefaultLabelFlagShorthand = "l"
	agentDefaultLabelFlagUsage     = "Default Label for this agent. Defaults to blank if not set." +
		" Alternatively, this can be set with the following environment variable: " + agentDefaultLabelEnvKey

	// log level.
	agentLogLevelFlagName  = "log-level"
	agentLogLevelEnvKey    = "ARIESD_LOG_LEVEL"
	agentLogLevelFlagUsage = "Log level." +
		" Possible values [INFO] [DEBUG] [ERROR] [WARNING] [CRITICAL] . Defaults to INFO if not set." +
		" Alternatively, this can be set with the following environment variable: " + agentLogLevelEnvKey

	// http resolver url flag.
	agentHTTPResolverFlagName      = "http-resolver-url"
	agentHTTPResolverEnvKey        = "ARIESD_HTTP_RESOLVER"
	agentHTTPResolverFlagShorthand = "r"
	agentHTTPResolverFlagUsage     = "HTTP binding DID resolver method and url. Values should be in `method@url` format." +
		" This flag can be repeated, allowing multiple http resolvers. Defaults to peer DID resolver if not set." +
		" Alternatively, this can be set with the following environment variable (in CSV format): " +
		agentHTTPResolverEnvKey

	// trustbloc domain url flag.
	agentTrustblocDomainFlagName      = "trustbloc-domain"
	agentTrustblocDomainEnvKey        = "ARIESD_TRUSTBLOC_DOMAIN"
	agentTrustblocDomainFlagShorthand = "d"
	agentTrustblocDomainFlagUsage     = "Trustbloc domain URL." +
		" Alternatively, this can be set with the following environment variable (in CSV format): " +
		agentTrustblocDomainEnvKey

	// EDV server URL flag.
	agentEDVServerURLFlagName  = "edv-server-url"
	agentEDVServerURLEnvKey    = "ARIESD_EDV_SERVER_URL"
	agentEDVServerURLFlagUsage = "EDV server URL." +
		" Alternatively, this can be set with the following environment variable (in CSV format): " +
		agentEDVServerURLEnvKey

	// trustbloc resolver url flag.
	agentTrustblocResolverFlagName  = "trustbloc-resolver"
	agentTrustblocResolverEnvKey    = "ARIESD_TRUSTBLOC_RESOLVER"
	agentTrustblocResolverFlagUsage = "Trustbloc resolver URL." +
		" Alternatively, this can be set with the following environment variable (in CSV format): " +
		agentTrustblocResolverEnvKey

	// outbound transport flag.
	agentOutboundTransportFlagName      = "outbound-transport"
	agentOutboundTransportEnvKey        = "ARIESD_OUTBOUND_TRANSPORT"
	agentOutboundTransportFlagShorthand = "o"
	agentOutboundTransportFlagUsage     = "Outbound transport type." +
		" This flag can be repeated, allowing for multiple transports." +
		" Possible values [http] [ws]. Defaults to http if not set." +
		" Alternatively, this can be set with the following environment variable: " + agentOutboundTransportEnvKey

	agentTLSCertFileFlagName      = "tls-cert-file"
	agentTLSCertFileEnvKey        = "TLS_CERT_FILE"
	agentTLSCertFileFlagShorthand = "c"
	agentTLSCertFileFlagUsage     = "tls certificate file." +
		" Alternatively, this can be set with the following environment variable: " + agentTLSCertFileEnvKey

	agentTLSKeyFileFlagName      = "tls-key-file"
	agentTLSKeyFileEnvKey        = "TLS_KEY_FILE"
	agentTLSKeyFileFlagShorthand = "k"
	agentTLSKeyFileFlagUsage     = "tls key file." +
		" Alternatively, this can be set with the following environment variable: " + agentTLSKeyFileEnvKey

	// inbound host url flag.
	agentInboundHostFlagName      = "inbound-host"
	agentInboundHostEnvKey        = "ARIESD_INBOUND_HOST"
	agentInboundHostFlagShorthand = "i"
	agentInboundHostFlagUsage     = "Inbound Host Name:Port. This is used internally to start the inbound server." +
		" Values should be in `scheme@url` format." +
		" This flag can be repeated, allowing to configure multiple inbound transports." +
		" Alternatively, this can be set with the following environment variable: " + agentInboundHostEnvKey

	// inbound host external url flag.
	agentInboundHostExternalFlagName      = "inbound-host-external"
	agentInboundHostExternalEnvKey        = "ARIESD_INBOUND_HOST_EXTERNAL"
	agentInboundHostExternalFlagShorthand = "e"
	agentInboundHostExternalFlagUsage     = "Inbound Host External Name:Port and values should be in `scheme@url` format" +
		" This is the URL for the inbound server as seen externally." +
		" If not provided, then the internal inbound host will be used here." +
		" This flag can be repeated, allowing to configure multiple inbound transports." +
		" Alternatively, this can be set with the following environment variable: " + agentInboundHostExternalEnvKey

	// auto accept flag.
	agentAutoAcceptFlagName  = "auto-accept"
	agentAutoAcceptEnvKey    = "ARIESD_AUTO_ACCEPT"
	agentAutoAcceptFlagUsage = "Auto accept requests." +
		" Possible values [true] [false]. Defaults to false if not set." +
		" Alternatively, this can be set with the following environment variable: " + agentAutoAcceptEnvKey

	// transport return route option flag.
	agentTransportReturnRouteFlagName  = "transport-return-route"
	agentTransportReturnRouteEnvKey    = "ARIESD_TRANSPORT_RETURN_ROUTE"
	agentTransportReturnRouteFlagUsage = "Transport Return Route option." +
		" Refer https://github.com/hyperledger/aries-framework-go/blob/8449c727c7c44f47ed7c9f10f35f0cd051dcb4e9/" +
		"pkg/framework/aries/framework.go#L165-L168." +
		" Alternatively, this can be set with the following environment variable: " + agentTransportReturnRouteEnvKey

	// remote JSON-LD context provider url flag.
	agentContextProviderFlagName  = "context-provider-url"
	agentContextProviderEnvKey    = "ARIESD_CONTEXT_PROVIDER_URL"
	agentContextProviderFlagUsage = "Remote context provider URL to get JSON-LD contexts from." +
		" This flag can be repeated, allowing setting up multiple context providers." +
		" Alternatively, this can be set with the following environment variable (in CSV format): " +
		agentContextProviderEnvKey

	// default verification key type flag.
	agentKeyTypeFlagName = "key-type"
	agentKeyTypeEnvKey   = "ARIESD_KEY_TYPE"
	agentKeyTypeUsage    = "Default key type supported by this agent." +
		" This flag sets the verification (and for DIDComm V1 encryption as well) key type used for key creation in the agent." + //nolint:lll
		" Alternatively, this can be set with the following environment variable: " +
		agentKeyTypeEnvKey

	// default key agreement type flag.
	agentKeyAgreementTypeFlagName = "key-agreement-type"
	agentKeyAgreementTypeEnvKey   = "ARIESD_KEY_AGREEMENT_TYPE"
	agentKeyAgreementTypeUsage    = "Default key agreement type supported by this agent." +
		" Default encryption (used in DIDComm V2) key type used for key agreement creation in the agent." +
		" Alternatively, this can be set with the following environment variable: " +
		agentKeyAgreementTypeEnvKey

	httpProtocol      = "http"
	websocketProtocol = "ws"

	databaseTypeMemOption     = "mem"
	databaseTypeCouchDBOption = "couchdb"
	databaseTypeMYSQLDBOption = "mysql"
	databaseTypeLevelDBOption = "leveldb"
	databaseTypeMongoDBOption = "mongodb"
)

var (
	errMissingHost = errors.New("host not provided")
	logger         = log.New("agent-sdk/agent-rest")
)

type agentParameters struct {
	server                                         server
	host, defaultLabel, transportReturnRoute       string
	tlsCertFile, tlsKeyFile                        string
	token                                          string
	trustblocDomain                                string
	edvServerURL                                   string
	trustblocResolver                              string
	webhookURLs, httpResolvers, outboundTransports []string
	inboundHostInternals, inboundHostExternals     []string
	contextProviderURLs                            []string
	autoAccept                                     bool
	msgHandler                                     command.MessageHandler
	dbParam                                        *dbParam
	keyType                                        string
	keyAgreementType                               string
}

type dbParam struct {
	dbType  string
	url     string
	prefix  string
	timeout uint64
}

// TODO (#47): Add support for EDV storage.
// nolint:gochecknoglobals
var supportedStorageProviders = map[string]func(url, prefix string) (storage.Provider, error){
	databaseTypeMemOption: func(_, _ string) (storage.Provider, error) { // nolint:unparam
		return mem.NewProvider(), nil
	},
	databaseTypeLevelDBOption: func(_, path string) (storage.Provider, error) { // nolint:unparam
		return leveldb.NewProvider(path), nil
	},
	databaseTypeCouchDBOption: func(url, prefix string) (storage.Provider, error) {
		return couchdb.NewProvider(url, couchdb.WithDBPrefix(prefix))
	},
	databaseTypeMYSQLDBOption: func(url, prefix string) (storage.Provider, error) {
		return mysql.NewProvider(url, mysql.WithDBPrefix(prefix))
	},
	databaseTypeMongoDBOption: func(url, prefix string) (storage.Provider, error) {
		return mongodb.NewProvider(url, mongodb.WithDBPrefix(prefix))
	},
}

type server interface {
	ListenAndServe(host string, router http.Handler, certFile, keyFile string) error
}

// HTTPServer represents an actual server implementation.
type HTTPServer struct{}

// ListenAndServe starts the server using the standard Go HTTP server implementation.
func (s *HTTPServer) ListenAndServe(host string, router http.Handler, certFile, keyFile string) error {
	if certFile != "" && keyFile != "" {
		return http.ListenAndServeTLS(host, certFile, keyFile, router)
	}

	return http.ListenAndServe(host, router)
}

// Cmd returns the Cobra start command.
func Cmd(server server) (*cobra.Command, error) {
	startCmd := createStartCMD(server)

	createFlags(startCmd)

	return startCmd, nil
}

func createStartCMD(server server) *cobra.Command { // nolint: funlen, gocyclo, gocognit
	return &cobra.Command{
		Use:   "start",
		Short: "Start an agent",
		Long:  `Start an Aries agent controller`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// log level
			logLevel, err := getUserSetVar(cmd, agentLogLevelFlagName, agentLogLevelEnvKey, true)
			if err != nil {
				return err
			}

			err = setLogLevel(logLevel)
			if err != nil {
				return err
			}

			host, err := getUserSetVar(cmd, agentHostFlagName, agentHostEnvKey, false)
			if err != nil {
				return err
			}

			token, err := getUserSetVar(cmd, agentTokenFlagName, agentTokenEnvKey, true)
			if err != nil {
				return err
			}

			inboundHosts, err := getUserSetVars(cmd, agentInboundHostFlagName, agentInboundHostEnvKey, true)
			if err != nil {
				return err
			}

			inboundHostExternals, err := getUserSetVars(cmd, agentInboundHostExternalFlagName,
				agentInboundHostExternalEnvKey, true)
			if err != nil {
				return err
			}

			dbParam, err := getDBParam(cmd)
			if err != nil {
				return err
			}

			defaultLabel, err := getUserSetVar(cmd, agentDefaultLabelFlagName, agentDefaultLabelEnvKey, true)
			if err != nil {
				return err
			}

			autoAccept, err := getAutoAcceptValue(cmd)
			if err != nil {
				return err
			}

			webhookURLs, err := getUserSetVars(cmd, agentWebhookFlagName, agentWebhookEnvKey, true)
			if err != nil {
				return err
			}

			httpResolvers, err := getUserSetVars(cmd, agentHTTPResolverFlagName, agentHTTPResolverEnvKey, true)
			if err != nil {
				return err
			}

			trustblocDomain, err := getUserSetVar(cmd, agentTrustblocDomainFlagName, agentTrustblocDomainEnvKey, true)
			if err != nil {
				return err
			}

			edvServerURL, err := getUserSetVar(cmd, agentEDVServerURLFlagName, agentEDVServerURLEnvKey, true)
			if err != nil {
				return err
			}

			trustblocResolver, err := getUserSetVar(cmd, agentTrustblocResolverFlagName, agentTrustblocResolverEnvKey, true)
			if err != nil {
				return err
			}

			outboundTransports, err := getUserSetVars(cmd, agentOutboundTransportFlagName,
				agentOutboundTransportEnvKey, true)
			if err != nil {
				return err
			}

			transportReturnRoute, err := getUserSetVar(cmd, agentTransportReturnRouteFlagName,
				agentTransportReturnRouteEnvKey, true)
			if err != nil {
				return err
			}

			tlsCertFile, err := getUserSetVar(cmd, agentTLSCertFileFlagName, agentTLSCertFileEnvKey, true)
			if err != nil {
				return err
			}

			tlsKeyFile, err := getUserSetVar(cmd, agentTLSKeyFileFlagName, agentTLSKeyFileEnvKey, true)
			if err != nil {
				return err
			}

			contextProviderURLs, err := getUserSetVars(cmd, agentContextProviderFlagName, agentContextProviderEnvKey, true)
			if err != nil {
				return err
			}

			keyType, err := getUserSetVar(cmd, agentKeyTypeFlagName, agentKeyTypeEnvKey, true)
			if err != nil {
				return err
			}

			keyAgreementType, err := getUserSetVar(cmd, agentKeyAgreementTypeFlagName, agentKeyAgreementTypeEnvKey, true)
			if err != nil {
				return err
			}

			parameters := &agentParameters{
				server:               server,
				host:                 host,
				token:                token,
				inboundHostInternals: inboundHosts,
				inboundHostExternals: inboundHostExternals,
				dbParam:              dbParam,
				defaultLabel:         defaultLabel,
				webhookURLs:          webhookURLs,
				httpResolvers:        httpResolvers,
				trustblocDomain:      trustblocDomain,
				edvServerURL:         edvServerURL,
				trustblocResolver:    trustblocResolver,
				outboundTransports:   outboundTransports,
				autoAccept:           autoAccept,
				transportReturnRoute: transportReturnRoute,
				contextProviderURLs:  contextProviderURLs,
				tlsCertFile:          tlsCertFile,
				tlsKeyFile:           tlsKeyFile,
				keyType:              keyType,
				keyAgreementType:     keyAgreementType,
			}

			return startAgent(parameters)
		},
	}
}

func getDBParam(cmd *cobra.Command) (*dbParam, error) {
	dbParam := &dbParam{}

	var err error

	dbParam.dbType, err = getUserSetVar(cmd, databaseTypeFlagName, databaseTypeEnvKey, false)
	if err != nil {
		return nil, err
	}

	dbParam.url, err = getUserSetVar(cmd, databaseURLFlagName, databaseURLEnvKey, true)
	if err != nil {
		return nil, err
	}

	dbParam.prefix, err = getUserSetVar(cmd, databasePrefixFlagName, databasePrefixEnvKey, true)
	if err != nil {
		return nil, err
	}

	dbTimeout, err := getUserSetVar(cmd, databaseTimeoutFlagName, databaseTimeoutEnvKey, true)
	if err != nil {
		return nil, err
	}

	if dbTimeout == "" || dbTimeout == "0" {
		dbTimeout = databaseTimeoutDefault
	}

	t, err := strconv.Atoi(dbTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db timeout %s: %w", dbTimeout, err)
	}

	dbParam.timeout = uint64(t)

	return dbParam, nil
}

func getAutoAcceptValue(cmd *cobra.Command) (bool, error) {
	v, err := getUserSetVar(cmd, agentAutoAcceptFlagName, agentAutoAcceptEnvKey, true)
	if err != nil {
		return false, err
	}

	if v == "" {
		return false, nil
	}

	return strconv.ParseBool(v)
}

func createFlags(startCmd *cobra.Command) { // nolint: funlen
	// agent host flag
	startCmd.Flags().StringP(agentHostFlagName, agentHostFlagShorthand, "", agentHostFlagUsage)

	// agent token flag
	startCmd.Flags().StringP(agentTokenFlagName, agentTokenFlagShorthand, "", agentTokenFlagUsage)

	// inbound host flag
	startCmd.Flags().StringSliceP(agentInboundHostFlagName, agentInboundHostFlagShorthand, []string{},
		agentInboundHostFlagUsage)

	// inbound external host flag
	startCmd.Flags().StringSliceP(agentInboundHostExternalFlagName, agentInboundHostExternalFlagShorthand,
		[]string{}, agentInboundHostExternalFlagUsage)

	// db type
	startCmd.Flags().StringP(databaseTypeFlagName, databaseTypeFlagShorthand, "", databaseTypeFlagUsage)

	// db url
	startCmd.Flags().StringP(databaseURLFlagName, databaseURLFlagShorthand, "", databaseURLFlagUsage)

	// db prefix
	startCmd.Flags().StringP(databasePrefixFlagName, databasePrefixFlagShorthand, "", databasePrefixFlagUsage)

	// webhook url flag
	startCmd.Flags().StringSliceP(agentWebhookFlagName, agentWebhookFlagShorthand, []string{}, agentWebhookFlagUsage)

	// log level
	startCmd.Flags().StringP(agentLogLevelFlagName, "", "", agentLogLevelFlagUsage)

	// http resolver url flag
	startCmd.Flags().StringSliceP(agentHTTPResolverFlagName, agentHTTPResolverFlagShorthand, []string{},
		agentHTTPResolverFlagUsage)

	// trustbloc domain url flag
	startCmd.Flags().StringP(agentTrustblocDomainFlagName, agentTrustblocDomainFlagShorthand, "",
		agentTrustblocDomainFlagUsage)

	// sds server url flag
	startCmd.Flags().StringP(agentEDVServerURLFlagName, "", "",
		agentEDVServerURLFlagUsage)

	// trustbloc resolver url flag
	startCmd.Flags().StringP(agentTrustblocResolverFlagName, "", "", agentTrustblocResolverFlagUsage)

	// agent default label flag
	startCmd.Flags().StringP(agentDefaultLabelFlagName, agentDefaultLabelFlagShorthand, "",
		agentDefaultLabelFlagUsage)

	// agent outbound transport flag
	startCmd.Flags().StringSliceP(agentOutboundTransportFlagName, agentOutboundTransportFlagShorthand, []string{},
		agentOutboundTransportFlagUsage)

	// auto accept flag
	startCmd.Flags().StringP(agentAutoAcceptFlagName, "", "", agentAutoAcceptFlagUsage)

	// transport return route option flag
	startCmd.Flags().StringP(agentTransportReturnRouteFlagName, "", "", agentTransportReturnRouteFlagUsage)

	// tls cert file
	startCmd.Flags().StringP(agentTLSCertFileFlagName,
		agentTLSCertFileFlagShorthand, "", agentTLSCertFileFlagUsage)

	// tls key file
	startCmd.Flags().StringP(agentTLSKeyFileFlagName,
		agentTLSKeyFileFlagShorthand, "", agentTLSKeyFileFlagUsage)

	// db timeout
	startCmd.Flags().StringP(databaseTimeoutFlagName, "", "", databaseTimeoutFlagUsage)

	// remote JSON-LD context provider url flag
	startCmd.Flags().StringSliceP(agentContextProviderFlagName, "", []string{}, agentContextProviderFlagUsage)

	// key types
	startCmd.Flags().StringP(agentKeyTypeFlagName, "", "", agentKeyTypeUsage)
	startCmd.Flags().StringP(agentKeyAgreementTypeFlagName, "", "", agentKeyAgreementTypeUsage)
}

func getUserSetVar(cmd *cobra.Command, flagName, envKey string, isOptional bool) (string, error) {
	if cmd.Flags().Changed(flagName) {
		value, err := cmd.Flags().GetString(flagName)
		if err != nil {
			return "", fmt.Errorf(flagName+" flag not found: %s", err)
		}

		return value, nil
	}

	value, isSet := os.LookupEnv(envKey)

	if isOptional || isSet {
		return value, nil
	}

	return "", errors.New("Neither " + flagName + " (command line flag) nor " + envKey +
		" (environment variable) have been set.")
}

func getUserSetVars(cmd *cobra.Command, flagName, envKey string, isOptional bool) ([]string, error) {
	if cmd.Flags().Changed(flagName) {
		value, err := cmd.Flags().GetStringSlice(flagName)
		if err != nil {
			return nil, fmt.Errorf(flagName+" flag not found: %s", err)
		}

		return value, nil
	}

	value, isSet := os.LookupEnv(envKey)

	var values []string

	if isSet {
		values = strings.Split(value, ",")
	}

	if isOptional || isSet {
		return values, nil
	}

	return nil, fmt.Errorf(" %s not set. "+
		"It must be set via either command line or environment variable", flagName)
}

func createVDRs(resolvers []string, trustblocDomain string) ([]vdr.VDR, error) {
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

	blocVDR, err := orb.New(nil,
		orb.WithDomain(trustblocDomain),
		orb.WithHTTPClient(http.DefaultClient),
		orb.WithVerifyResolutionResultType(orb.None),
	)
	if err != nil {
		return nil, err
	}

	VDRs = append(VDRs, blocVDR)

	return VDRs, nil
}

func getOutboundTransportOpts(outboundTransports []string) ([]aries.Option, error) {
	var opts []aries.Option

	var transports []transport.OutboundTransport

	for _, outboundTransport := range outboundTransports {
		switch outboundTransport {
		case httpProtocol:
			outbound, err := arieshttp.NewOutbound(arieshttp.WithOutboundHTTPClient(&http.Client{}))
			if err != nil {
				return nil, fmt.Errorf("http outbound transport initialization failed: %w", err)
			}

			transports = append(transports, outbound)
		case websocketProtocol:
			transports = append(transports, ws.NewOutbound())
		default:
			return nil, fmt.Errorf("outbound transport [%s] not supported", outboundTransport)
		}
	}

	if len(transports) > 0 {
		opts = append(opts, aries.WithOutboundTransports(transports...))
	}

	return opts, nil
}

func getInboundTransportOpts(inboundHostInternals, inboundHostExternals []string, certFile,
	keyFile string) ([]aries.Option, error) {
	internalHost, err := getInboundSchemeToURLMap(inboundHostInternals)
	if err != nil {
		return nil, fmt.Errorf("inbound internal host : %w", err)
	}

	externalHost, err := getInboundSchemeToURLMap(inboundHostExternals)
	if err != nil {
		return nil, fmt.Errorf("inbound external host : %w", err)
	}

	var opts []aries.Option

	for scheme, host := range internalHost {
		switch scheme {
		case httpProtocol:
			opts = append(opts, defaults.WithInboundHTTPAddr(host, externalHost[scheme], certFile, keyFile))
		case websocketProtocol:
			opts = append(opts, defaults.WithInboundWSAddr(host, externalHost[scheme], certFile, keyFile))
		default:
			return nil, fmt.Errorf("inbound transport [%s] not supported", scheme)
		}
	}

	return opts, nil
}

func getInboundSchemeToURLMap(schemeHostStr []string) (map[string]string, error) {
	const validSliceLen = 2

	schemeHostMap := make(map[string]string)

	for _, schemeHost := range schemeHostStr {
		schemeHostSlice := strings.Split(schemeHost, "@")
		if len(schemeHostSlice) != validSliceLen {
			return nil, fmt.Errorf("invalid inbound host option: Use scheme@url to pass the option")
		}

		schemeHostMap[schemeHostSlice[0]] = schemeHostSlice[1]
	}

	return schemeHostMap, nil
}

func setLogLevel(logLevel string) error {
	if logLevel != "" {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("failed to parse log level '%s' : %w", logLevel, err)
		}

		log.SetLevel("", level)

		logger.Infof("logger level set to %s", logLevel)
	}

	return nil
}

func validateAuthorizationBearerToken(w http.ResponseWriter, r *http.Request, token string) bool {
	actHdr := r.Header.Get("Authorization")
	expHdr := "Bearer " + token

	if subtle.ConstantTimeCompare([]byte(actHdr), []byte(expHdr)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorised.\n")) // nolint:gosec,errcheck

		return false
	}

	return true
}

func authorizationMiddleware(token string) mux.MiddlewareFunc {
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if validateAuthorizationBearerToken(w, r, token) {
				next.ServeHTTP(w, r)
			}
		})
	}

	return middleware
}

func startAgent(parameters *agentParameters) error {
	if parameters.host == "" {
		return errMissingHost
	}

	// set message handler
	parameters.msgHandler = msghandler.NewRegistrar()

	ctx, err := createAriesAgent(parameters)
	if err != nil {
		return err
	}

	// get all HTTP REST API handlers available for controller API
	handlers, err := controller.GetRESTHandlers(ctx, controller.WithWebhookURLs(parameters.webhookURLs...),
		controller.WithDefaultLabel(parameters.defaultLabel), controller.WithAutoAccept(parameters.autoAccept),
		controller.WithMessageHandler(parameters.msgHandler))
	if err != nil {
		return fmt.Errorf("failed to start aries agent rest on port [%s], failed to get rest service api :  %w",
			parameters.host, err)
	}

	sdkHandlers, err := sdkcontroller.GetRESTHandlers(ctx, sdkcontroller.WithBlocDomain(parameters.trustblocDomain),
		sdkcontroller.WithMessageHandler(parameters.msgHandler))
	if err != nil {
		return fmt.Errorf("failed to start sdk agent rest on port [%s], failed to get rest service api:  %w",
			parameters.host, err)
	}

	for i := range sdkHandlers {
		handlers = append(handlers, sdkHandlers[i])
	}

	router := mux.NewRouter()

	if parameters.token != "" {
		router.Use(authorizationMiddleware(parameters.token))
	}

	for _, handler := range handlers {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	logger.Infof("Starting aries agent rest on host [%s]", parameters.host)
	// start server on given port and serve using given handlers
	handler := cors.New(
		cors.Options{
			AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodHead},
			AllowedHeaders: []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization"},
		},
	).Handler(router)

	err = parameters.server.ListenAndServe(parameters.host, handler, parameters.tlsCertFile, parameters.tlsKeyFile)
	if err != nil {
		return fmt.Errorf("failed to start aries agent rest on port [%s], cause:  %w", parameters.host, err)
	}

	return nil
}

var (
	//nolint:gochecknoglobals
	keyTypes = map[string]kms.KeyType{
		"ed25519":           kms.ED25519Type,
		"ecdsap256ieee1363": kms.ECDSAP256TypeIEEEP1363,
		"ecdsap256der":      kms.ECDSAP256TypeDER,
		"ecdsap384ieee1363": kms.ECDSAP384TypeIEEEP1363,
		"ecdsap384der":      kms.ECDSAP384TypeDER,
		"ecdsap521ieee1363": kms.ECDSAP521TypeIEEEP1363,
		"ecdsap521der":      kms.ECDSAP521TypeDER,
	}

	//nolint:gochecknoglobals
	keyAgreementTypes = map[string]kms.KeyType{
		"x25519kw": kms.X25519ECDHKWType,
		"p256kw":   kms.NISTP256ECDHKWType,
		"p384kw":   kms.NISTP384ECDHKWType,
		"p521kw":   kms.NISTP521ECDHKWType,
	}
)

func createAriesAgent(parameters *agentParameters) (*context.Provider, error) { // nolint: funlen,gocyclo
	var opts []aries.Option

	storePro, err := createStoreProviders(parameters)
	if err != nil {
		return nil, err
	}

	opts = append(opts, aries.WithStoreProvider(storePro))

	if parameters.transportReturnRoute != "" {
		opts = append(opts, aries.WithTransportReturnRoute(parameters.transportReturnRoute))
	}

	inboundTransportOpt, err := getInboundTransportOpts(parameters.inboundHostInternals,
		parameters.inboundHostExternals, parameters.tlsCertFile, parameters.tlsKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to start aries agent rest on port [%s], failed to inbound tranpsort opt : %w",
			parameters.host, err)
	}

	opts = append(opts, inboundTransportOpt...)

	VDRs, err := createVDRs(parameters.httpResolvers, parameters.trustblocDomain)
	if err != nil {
		return nil, err
	}

	for i := range VDRs {
		opts = append(opts, aries.WithVDR(VDRs[i]))
	}

	outboundTransportOpts, err := getOutboundTransportOpts(parameters.outboundTransports)
	if err != nil {
		return nil, fmt.Errorf("failed to start aries agent rest on port [%s], failed to outbound transport opts : %w",
			parameters.host, err)
	}

	opts = append(opts, outboundTransportOpts...)
	opts = append(opts, aries.WithMessageServiceProvider(parameters.msgHandler))

	if len(parameters.contextProviderURLs) > 0 {
		opts = append(opts, aries.WithJSONLDContextProviderURL(parameters.contextProviderURLs...))
	}

	if kt, ok := keyTypes[parameters.keyType]; ok {
		opts = append(opts, aries.WithKeyType(kt))
	}

	if kat, ok := keyAgreementTypes[parameters.keyAgreementType]; ok {
		opts = append(opts, aries.WithKeyAgreementType(kat))
	}

	framework, err := aries.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to start aries agent rest on port [%s], failed to initialize framework :  %w",
			parameters.host, err)
	}

	ctx, err := framework.Context()
	if err != nil {
		return nil, fmt.Errorf("failed to start aries agent rest on port [%s], failed to get aries context : %w",
			parameters.host, err)
	}

	return ctx, nil
}

func createStoreProviders(parameters *agentParameters) (storage.Provider, error) {
	provider, supported := supportedStorageProviders[parameters.dbParam.dbType]
	if !supported {
		return nil, fmt.Errorf("key database type not set to a valid type." +
			" run start --help to see the available options")
	}

	var store storage.Provider

	err := backoff.RetryNotify(
		func() error {
			var openErr error
			store, openErr = provider(parameters.dbParam.url, parameters.dbParam.prefix)

			return openErr
		},
		backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), parameters.dbParam.timeout),
		func(retryErr error, t time.Duration) {
			logger.Warnf(
				"failed to connect to storage, will sleep for %s before trying again : %s\n",
				t, retryErr)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to storage at %s: %w", parameters.dbParam.url, err)
	}

	return store, nil
}
