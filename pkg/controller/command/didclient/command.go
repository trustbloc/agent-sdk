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
	"strings"

	"github.com/hyperledger/aries-framework-go-ext/component/vdr/orb"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/sidetree/doc"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	mediatorservice "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	ariesjose "github.com/hyperledger/aries-framework-go/pkg/doc/jose"
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
	// CreateTrustBlocDIDCommandMethod command method.
	CreateTrustBlocDIDCommandMethod = "CreateTrustBlocDID"
	// CreatePeerDIDCommandMethod command method.
	CreatePeerDIDCommandMethod = "CreatePeerDID"
	// log constants.
	successString = "success"

	didCommServiceType = "did-communication"

	// ed25519KeyType defines ed25119 key type.
	ed25519KeyType = "Ed25519"

	// p256KeyType EC P-256 key type.
	p256KeyType = "P256"

	didParts = 3
)

const (
	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + command.DIDClient)

	// CreateDIDErrorCode is typically a code for create did errors.
	CreateDIDErrorCode

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
}

// mediatorClient is client interface for mediator.
type mediatorClient interface {
	Register(connectionID string) error
	GetConfig(connID string) (*mediatorservice.Config, error)
}

// New returns new DID Exchange controller command instance.
func New(domain, didAnchorOrigin string, p Provider) (*Command, error) {
	client, err := orb.New(nil, orb.WithDomain(domain))
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
		cmdutil.NewCommandHandler(CommandName, CreateTrustBlocDIDCommandMethod, c.CreateTrustBlocDID),
		cmdutil.NewCommandHandler(CommandName, CreatePeerDIDCommandMethod, c.CreatePeerDID),
	}
}

// CreateTrustBlocDID creates a new trust bloc DID.
func (c *Command) CreateTrustBlocDID(rw io.Writer, req io.Reader) command.Error { //nolint: funlen,gocyclo
	var request CreateBlocDIDRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	didDoc := did.Doc{}

	var didMethodOpt []vdr.DIDMethodOption

	for _, v := range request.PublicKeys {
		value, decodeErr := base64.RawURLEncoding.DecodeString(v.Value)
		if decodeErr != nil {
			logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, decodeErr.Error())

			return command.NewExecuteError(CreateDIDErrorCode, decodeErr)
		}

		k, errGet := getKey(v.KeyType, value)
		if errGet != nil {
			logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, errGet.Error())

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

		jwk, errJWK := ariesjose.JWKFromKey(k)
		if errJWK != nil {
			logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, errJWK.Error())

			return command.NewExecuteError(CreateDIDErrorCode, errJWK)
		}

		vm, errVM := did.NewVerificationMethodFromJWK(v.ID, v.Type, "", jwk)
		if errVM != nil {
			logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, errVM.Error())

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
				logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod,
					fmt.Sprintf("public key purpose %s not supported", p))

				return command.NewExecuteError(CreateDIDErrorCode,
					fmt.Errorf("public key purpose %s not supported", p))
			}
		}
	}

	didMethodOpt = append(didMethodOpt, vdr.WithOption(orb.AnchorOriginOpt, c.didAnchorOrigin))

	docResolution, err := c.didBlocClient.Create(&didDoc, didMethodOpt...)
	if err != nil {
		logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	bytes, err := docResolution.DIDDocument.JSONBytes()
	if err != nil {
		logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	didSplit := strings.Split(docResolution.DIDDocument.ID, ":")

	if len(didSplit) < didParts {
		return command.NewExecuteError(CreateDIDErrorCode,
			fmt.Errorf("did parts less than %d", didParts))
	}

	discoverableDID := strings.ReplaceAll(string(bytes),
		fmt.Sprintf("%s:%s", didSplit[0], didSplit[1]),
		fmt.Sprintf("%s:%s:%s", didSplit[0], didSplit[1], c.domain))

	command.WriteNillableResponse(rw, &CreateDIDResponse{
		DID: []byte(discoverableDID),
	}, logger)

	logutil.LogDebug(logger, CommandName, CreateTrustBlocDIDCommandMethod, successString)

	return nil
}

func getKey(keyType string, value []byte) (interface{}, error) {
	switch keyType {
	case ed25519KeyType:
		return ed25519.PublicKey(value), nil
	case p256KeyType:
		x, y := elliptic.Unmarshal(elliptic.P256(), value)

		return &ecdsa.PublicKey{X: x, Y: y, Curve: elliptic.P256()}, nil
	default:
		return nil, fmt.Errorf("invalid key type: %s", keyType)
	}
}

// CreatePeerDID creates a new peer DID.
func (c *Command) CreatePeerDID(rw io.Writer, req io.Reader) command.Error { //nolint: funlen
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

	bytes, err := docResolution.DIDDocument.JSONBytes()
	if err != nil {
		logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	command.WriteNillableResponse(rw, &CreateDIDResponse{
		DID: bytes,
	}, logger)

	logutil.LogDebug(logger, CommandName, CreateTrustBlocDIDCommandMethod, successString)

	return nil
}
