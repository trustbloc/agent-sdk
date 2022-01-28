/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package didclient provides did commands.
package didclient

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/hyperledger/aries-framework-go-ext/component/vdr/orb"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/sidetree/doc"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/primitive/bbs12381g2pub"
	mediatorservice "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	jwk2 "github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk/jwksupport"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/peer"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/logutil"
)

var logger = log.New("agent-sdk-didclient")

const (
	// CommandName package command name.
	CommandName = "didclient"
	// CreateOrbDIDCommandMethod command method.
	CreateOrbDIDCommandMethod = "CreateOrbDID"
	// ResolveOrbDIDCommandMethod command method.
	ResolveOrbDIDCommandMethod = "ResolveOrbDID"
	// CreatePeerDIDCommandMethod command method.
	CreatePeerDIDCommandMethod = "CreatePeerDID"
	// log constants.
	successString = "success"

	didCommServiceType = "did-communication"

	// ed25519KeyType defines ed25119 key type.
	ed25519KeyType = "ed25519"

	// p256KeyType EC P-256 key type.
	p256KeyType = "ecdsap256ieeep1363"

	// p384KeyType EC P-384 key type.
	p384KeyType = "ecdsap384ieeep1363"

	// BLS12381G2KeyType BLS12381G2 key type.
	BLS12381G2KeyType = "bls12381g2"

	// x25519ECDHKW key agreement type.
	x25519ECDHKW = "x25519ecdhkw"

	// NIST P curved key agreement types.
	p256ecdhkw = "nistp256ecdhkw"
	p384ecdhkw = "nistp384ecdhkw"
	p521ecdhkw = "nistp521ecdhkw"
)

const (
	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + command.DIDClient)

	// CreateDIDErrorCode is typically a code for create did errors.
	CreateDIDErrorCode

	// ResolveDIDErrorCode is typically a code for resolve did errors.
	ResolveDIDErrorCode

	// errors.
	errInvalidRouterConnectionID = "invalid router connection ID"
	errMissingDIDCommServiceType = "did document missing '%s' service type"
	errFailedToRegisterDIDRecKey = "failed to register did doc recipient key : %w"
)

// Provider describes dependencies for the client.
type Provider interface {
	VDRegistry() vdr.Registry
	Service(id string) (interface{}, error)
	KMS() kms.KeyManager
}

type didBlocClient interface {
	Create(did *did.Doc, opts ...vdr.DIDMethodOption) (*did.DocResolution, error)
	Read(id string, opts ...vdr.DIDMethodOption) (*did.DocResolution, error)
}

// mediatorClient is client interface for mediator.
type mediatorClient interface {
	Register(connectionID string) error
	GetConfig(connID string) (*mediatorservice.Config, error)
}

// New returns new DID Exchange controller command instance.
func New(domain, didAnchorOrigin, token string, unanchoredDIDMaxLifeTime int, p Provider) (*Command, error) {
	orbOpts := make([]orb.Option, 0)

	if unanchoredDIDMaxLifeTime > 0 {
		orbOpts = append(orbOpts, orb.WithUnanchoredMaxLifeTime(time.Duration(unanchoredDIDMaxLifeTime)*time.Second))
	}

	orbOpts = append(orbOpts, orb.WithDomain(domain), orb.WithAuthToken(token), orb.WithHTTPClient(http.DefaultClient))

	client, err := orb.New(nil, orbOpts...)
	if err != nil {
		return nil, err
	}

	mClient, err := mediator.New(p)
	if err != nil {
		return nil, err
	}

	var s interface{}

	s, err = p.Service(mediatorservice.Coordination)
	if err != nil {
		return nil, err
	}

	mediatorSvc, ok := s.(mediatorservice.ProtocolService)
	if !ok {
		return nil, errors.New("cast service to route service failed")
	}

	return &Command{
		didBlocClient:   client,
		domain:          domain,
		vdrRegistry:     p.VDRegistry(),
		mediatorClient:  mClient,
		mediatorSvc:     mediatorSvc,
		keyManager:      p.KMS(),
		didAnchorOrigin: didAnchorOrigin,
	}, nil
}

// Command is controller command for DID Exchange.
type Command struct {
	didBlocClient   didBlocClient
	domain          string
	vdrRegistry     vdr.Registry
	mediatorClient  mediatorClient
	mediatorSvc     mediatorservice.ProtocolService
	keyManager      kms.KeyManager
	didAnchorOrigin string
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []command.Handler {
	return []command.Handler{
		cmdutil.NewCommandHandler(CommandName, CreateOrbDIDCommandMethod, c.CreateOrbDID),
		cmdutil.NewCommandHandler(CommandName, CreatePeerDIDCommandMethod, c.CreatePeerDID),
		cmdutil.NewCommandHandler(CommandName, ResolveOrbDIDCommandMethod, c.ResolveOrbDID),
	}
}

// ResolveOrbDID resolve orb DID.
func (c *Command) ResolveOrbDID(rw io.Writer, req io.Reader) command.Error {
	var request ResolveOrbDIDRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, ResolveOrbDIDCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	docResolution, errRead := c.didBlocClient.Read(request.DID)
	if errRead != nil {
		logutil.LogError(logger, CommandName, ResolveOrbDIDCommandMethod, errRead.Error())

		return command.NewExecuteError(ResolveDIDErrorCode, errRead)
	}

	bytes, err := docResolution.JSONBytes()
	if err != nil {
		logutil.LogError(logger, CommandName, ResolveOrbDIDCommandMethod, err.Error())

		return command.NewExecuteError(ResolveDIDErrorCode, err)
	}

	logutil.LogDebug(logger, CommandName, ResolveOrbDIDCommandMethod, successString)

	if _, err := rw.Write(bytes); err != nil {
		logger.Errorf(err.Error())
	}

	return nil
}

// CreateOrbDID creates a new orb DID.
func (c *Command) CreateOrbDID(rw io.Writer, req io.Reader) command.Error { //nolint: funlen,gocyclo,gocognit
	var request CreateOrbDIDRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, fmt.Sprintf("request: %+v", request))

	didDoc := did.Doc{}

	didcommv2Servicetype := "DIDCommMessaging"
	if request.DIDcommServiceType != "" {
		didcommv2Servicetype = request.DIDcommServiceType
	}

	serviceID := "sidetree"
	if request.ServiceID != "" {
		serviceID = request.ServiceID
	}

	serviceEndpoint := "https://testnet.orb.local"
	if request.ServiceEndpoint != "" {
		serviceEndpoint = request.ServiceEndpoint
	}

	logutil.LogDebug(logger, CommandName, CreateOrbDIDCommandMethod, fmt.Sprintf("request.RoutersKeyAgrIDS: %+v",
		request.RoutersKeyAgrIDS))

	var routerKeys []string
	routerKeys = append(routerKeys, request.RoutersKeyAgrIDS...)

	logutil.LogDebug(logger, CommandName, CreateOrbDIDCommandMethod, fmt.Sprintf("routerKeys: %+v", routerKeys))

	didDoc.Service = []did.Service{{
		ID:              serviceID,
		Type:            didcommv2Servicetype,
		ServiceEndpoint: serviceEndpoint,
		RoutingKeys:     routerKeys,
	}}

	var didMethodOpt []vdr.DIDMethodOption

	for _, v := range request.PublicKeys {
		value, decodeErr := base64.RawURLEncoding.DecodeString(v.Value)
		if decodeErr != nil {
			logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, decodeErr.Error())

			return command.NewExecuteError(CreateDIDErrorCode, decodeErr)
		}

		k, errGet := getKey(v.KeyType, value)
		if errGet != nil {
			logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, errGet.Error())

			return command.NewExecuteError(CreateDIDErrorCode, errGet)
		}

		if v.Recovery {
			didMethodOpt = append(didMethodOpt, vdr.WithOption(orb.RecoveryPublicKeyOpt, k))

			continue
		}

		if v.Update {
			didMethodOpt = append(didMethodOpt, vdr.WithOption(orb.UpdatePublicKeyOpt, k))

			continue
		}

		var (
			jwk    *jwk2.JWK
			errJWK error
		)

		//nolint:gocritic,nestif
		if strings.EqualFold(v.KeyType, x25519ECDHKW) {
			jwk, errJWK = jwksupport.JWKFromX25519Key(k.(*crypto.PublicKey).X)
			if errJWK != nil {
				logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, errJWK.Error())

				return command.NewExecuteError(CreateDIDErrorCode, errJWK)
			}
		} else if strings.EqualFold(v.KeyType, p256ecdhkw) || strings.EqualFold(v.KeyType, p384ecdhkw) ||
			strings.EqualFold(v.KeyType, p521ecdhkw) {
			pubKey, ok := k.(*crypto.PublicKey)
			if !ok {
				logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod,
					fmt.Sprintf("key '%+v' is not NIST P ECDH KW type", k))

				return command.NewExecuteError(CreateDIDErrorCode, fmt.Errorf("key '%+v' is not NIST P ECDH KW type", k))
			}

			ecdsaKey := &ecdsa.PublicKey{
				X:     new(big.Int).SetBytes(pubKey.X),
				Y:     new(big.Int).SetBytes(pubKey.Y),
				Curve: getCurve(pubKey.Curve),
			}

			jwk, errJWK = jwksupport.JWKFromKey(ecdsaKey)
			if errJWK != nil {
				logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, errJWK.Error())

				return command.NewExecuteError(CreateDIDErrorCode, fmt.Errorf("JWKFromKey() jwk: %+v, ecdsa key: "+
					"%+v, error: %w", jwk, ecdsaKey, errJWK))
			}
		} else {
			jwk, errJWK = jwksupport.JWKFromKey(k)
			if errJWK != nil {
				logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, errJWK.Error())

				return command.NewExecuteError(CreateDIDErrorCode, errJWK)
			}
		}

		vm, errVM := did.NewVerificationMethodFromJWK(v.ID, v.Type, "", jwk)
		if errVM != nil {
			logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, errVM.Error())

			return command.NewExecuteError(CreateDIDErrorCode, errVM)
		}

		for _, p := range v.Purposes {
			switch p {
			case doc.KeyPurposeAuthentication:
				didDoc.Authentication = append(didDoc.Authentication,
					*did.NewReferencedVerification(vm, did.Authentication))
			case doc.KeyPurposeAssertionMethod:
				didDoc.AssertionMethod = append(didDoc.AssertionMethod,
					*did.NewReferencedVerification(vm, did.AssertionMethod))
			case doc.KeyPurposeKeyAgreement:
				didDoc.KeyAgreement = append(didDoc.KeyAgreement,
					*did.NewReferencedVerification(vm, did.KeyAgreement))
			case doc.KeyPurposeCapabilityDelegation:
				didDoc.CapabilityDelegation = append(didDoc.CapabilityDelegation,
					*did.NewReferencedVerification(vm, did.CapabilityDelegation))
			case doc.KeyPurposeCapabilityInvocation:
				didDoc.CapabilityInvocation = append(didDoc.CapabilityInvocation,
					*did.NewReferencedVerification(vm, did.CapabilityInvocation))
			default:
				logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod,
					fmt.Sprintf("public key purpose %s not supported", p))

				return command.NewExecuteError(CreateDIDErrorCode,
					fmt.Errorf("public key purpose %s not supported", p))
			}
		}
	}

	didMethodOpt = append(didMethodOpt, vdr.WithOption(orb.AnchorOriginOpt, c.didAnchorOrigin))

	logutil.LogDebug(logger, CommandName, CreateOrbDIDCommandMethod, fmt.Sprintf("about to create DID Doc, "+
		"keyAgreements: %+v", didDoc.KeyAgreement))

	docResolution, err := c.didBlocClient.Create(&didDoc, didMethodOpt...)
	if err != nil {
		logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	logutil.LogDebug(logger, CommandName, CreateOrbDIDCommandMethod, fmt.Sprintf("ORB DID Doc crated: %+v",
		docResolution.DIDDocument))

	// add all keyAgreements to router connections
	for _, val := range docResolution.DIDDocument.KeyAgreement {
		for _, rConn := range request.RouterConnections {
			err = mediatorservice.AddKeyToRouter(c.mediatorSvc, rConn, val.VerificationMethod.ID)
			if err != nil {
				logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, err.Error())

				return command.NewExecuteError(CreateDIDErrorCode, fmt.Errorf(errFailedToRegisterDIDRecKey+
					" for KeyAgreement ID %v, connection: %v", err, val.VerificationMethod.ID, rConn))
			}

			logutil.LogDebug(logger, CommandName, CreateOrbDIDCommandMethod, fmt.Sprintf("added keyAgreements ID %v"+
				" to router connection: %+v", val.VerificationMethod.ID, rConn))
		}
	}

	bytes, err := docResolution.JSONBytes()
	if err != nil {
		logutil.LogError(logger, CommandName, CreateOrbDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	logutil.LogDebug(logger, CommandName, CreateOrbDIDCommandMethod, successString)

	if _, err := rw.Write(bytes); err != nil {
		logger.Errorf(err.Error())
	}

	return nil
}

func getCurve(crv string) elliptic.Curve {
	c := elliptic.P256()

	switch crv {
	case "NIST_P256", "P-256":
	case "NIST_P384", "P-384":
		c = elliptic.P384()
	case "NIST_P521", "P-521":
		c = elliptic.P521()
	}

	return c
}

func getKey(keyType string, value []byte) (interface{}, error) {
	switch strings.ToLower(keyType) {
	case ed25519KeyType:
		return ed25519.PublicKey(value), nil
	case p256KeyType:
		x, y := elliptic.Unmarshal(elliptic.P256(), value)

		return &ecdsa.PublicKey{X: x, Y: y, Curve: elliptic.P256()}, nil
	case p384KeyType:
		x, y := elliptic.Unmarshal(elliptic.P384(), value)

		return &ecdsa.PublicKey{X: x, Y: y, Curve: elliptic.P384()}, nil
	case BLS12381G2KeyType:
		return bbs12381g2pub.UnmarshalPublicKey(value)
	case x25519ECDHKW, p256ecdhkw, p384ecdhkw, p521ecdhkw:
		pubKey := &crypto.PublicKey{}

		err := json.Unmarshal(value, pubKey)
		if err != nil {
			return nil, fmt.Errorf("unmarshal key type: %s, value: %s failed: %w", keyType, value, err)
		}

		return pubKey, nil
	default:
		return nil, fmt.Errorf("invalid key type: %s", keyType)
	}
}

// CreatePeerDID creates a new peer DID.
func (c *Command) CreatePeerDID(rw io.Writer, req io.Reader) command.Error { //nolint: funlen,gocyclo
	var request CreatePeerDIDRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	if request.RouterConnectionID == "" {
		logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, errInvalidRouterConnectionID)

		return command.NewValidationError(InvalidRequestErrorCode, fmt.Errorf(errInvalidRouterConnectionID))
	}

	config, err := c.mediatorClient.GetConfig(request.RouterConnectionID)
	if err != nil {
		logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	// TODO - key type should be configurable
	keyID, keyBytes, err := c.keyManager.CreateAndExportPubKeyBytes(kms.ED25519Type)
	if err != nil {
		logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	docResolution, err := c.vdrRegistry.Create(
		peer.DIDMethod,
		&did.Doc{
			Service: []did.Service{{
				ServiceEndpoint: config.Endpoint(),
				RoutingKeys:     config.Keys(),
			}},
			VerificationMethod: []did.VerificationMethod{*did.NewVerificationMethodFromBytes(
				"#"+keyID,
				"Ed25519VerificationKey2018",
				"",
				keyBytes,
			)},
		},
	)
	if err != nil {
		logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	didSvc, ok := did.LookupService(docResolution.DIDDocument, didCommServiceType)
	if !ok {
		logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, errMissingDIDCommServiceType)

		return command.NewExecuteError(CreateDIDErrorCode, fmt.Errorf(errMissingDIDCommServiceType, didCommServiceType))
	}

	for _, val := range didSvc.RecipientKeys {
		err = mediatorservice.AddKeyToRouter(c.mediatorSvc, request.RouterConnectionID, val)

		if err != nil {
			logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, err.Error())

			return command.NewExecuteError(CreateDIDErrorCode, fmt.Errorf(errFailedToRegisterDIDRecKey, err))
		}
	}

	bytes, err := docResolution.JSONBytes()
	if err != nil {
		logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	logutil.LogDebug(logger, CommandName, CreatePeerDIDCommandMethod, successString)

	if _, err := rw.Write(bytes); err != nil {
		logger.Errorf(err.Error())
	}

	return nil
}
