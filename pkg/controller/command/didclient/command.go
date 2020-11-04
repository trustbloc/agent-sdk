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

	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	mediatorservice "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/trustbloc/edge-core/pkg/log"
	didclient "github.com/trustbloc/trustbloc-did-method/pkg/did"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
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
	saveDIDCommandMethod       = "SaveDID"
	// log constants.
	successString = "success"

	didCommServiceType = "did-communication"
)

const (
	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + command.DIDClient)

	// CreateDIDErrorCode is typically a code for create did errors.
	CreateDIDErrorCode

	// errors.
	errDecodeDIDDocDataErrMsg    = "failure while decoding DID data"
	errStoreDIDDocErrMsg         = "failure while storing DID document in SDS"
	errInvalidRouterConnectionID = "invalid router connection ID"
	errMissingDIDCommServiceType = "did document missing '%s' service type"
	errFailedToRegisterDIDRecKey = "failed to register did doc recipient key : %w"
)

// Provider describes dependencies for the client.
type Provider interface {
	VDRegistry() vdr.Registry
	Service(id string) (interface{}, error)
}

type didBlocClient interface {
	CreateDID(domain string, opts ...didclient.CreateDIDOption) (*did.Doc, error)
}

// mediatorClient is client interface for mediator.
type mediatorClient interface {
	Register(connectionID string) error
	GetConfig(connID string) (*mediatorservice.Config, error)
}

// New returns new DID Exchange controller command instance.
func New(domain string, sdsComm *sdscomm.SDSComm, p Provider) (*Command, error) {
	client := didclient.New()

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
		didBlocClient:  client,
		domain:         domain,
		sdsComm:        sdsComm,
		vdrRegistry:    p.VDRegistry(),
		mediatorClient: mClient,
		mediatorSvc:    mediatorSvc,
	}, nil
}

// Command is controller command for DID Exchange.
type Command struct {
	didBlocClient  didBlocClient
	domain         string
	sdsComm        *sdscomm.SDSComm
	vdrRegistry    vdr.Registry
	mediatorClient mediatorClient
	mediatorSvc    mediatorservice.ProtocolService
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []command.Handler {
	return []command.Handler{
		cmdutil.NewCommandHandler(CommandName, CreateTrustBlocDIDCommandMethod, c.CreateTrustBlocDID),
		cmdutil.NewCommandHandler(CommandName, CreatePeerDIDCommandMethod, c.CreatePeerDID),
		cmdutil.NewCommandHandler(CommandName, saveDIDCommandMethod, c.SaveDID),
	}
}

// CreateTrustBlocDID creates a new trust bloc DID.
func (c *Command) CreateTrustBlocDID(rw io.Writer, req io.Reader) command.Error { //nolint: funlen
	var request CreateBlocDIDRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	var opts []didclient.CreateDIDOption

	for _, v := range request.PublicKeys {
		value, decodeErr := base64.RawURLEncoding.DecodeString(v.Value)
		if decodeErr != nil {
			logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, decodeErr.Error())

			return command.NewExecuteError(CreateDIDErrorCode, decodeErr)
		}

		if v.Recovery {
			k, recoverKeyErr := getKey(v.KeyType, value)
			if recoverKeyErr != nil {
				logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, recoverKeyErr.Error())

				return command.NewExecuteError(CreateDIDErrorCode, recoverKeyErr)
			}

			opts = append(opts, didclient.WithRecoveryPublicKey(k))

			continue
		}

		if v.Update {
			k, updateKeyErr := getKey(v.KeyType, value)
			if updateKeyErr != nil {
				logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, updateKeyErr.Error())

				return command.NewExecuteError(CreateDIDErrorCode, updateKeyErr)
			}

			opts = append(opts, didclient.WithUpdatePublicKey(k))

			continue
		}

		opts = append(opts, didclient.WithPublicKey(&didclient.PublicKey{
			ID: v.ID, Type: v.Type, Encoding: v.Encoding,
			KeyType: v.KeyType, Purposes: v.Purposes, Value: value,
		}))
	}

	didDoc, err := c.didBlocClient.CreateDID(c.domain, opts...)
	if err != nil {
		logutil.LogError(logger, CommandName, CreateTrustBlocDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	bytes, err := didDoc.JSONBytes()
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

func getKey(keyType string, value []byte) (interface{}, error) {
	switch keyType {
	case didclient.Ed25519KeyType:
		return ed25519.PublicKey(value), nil
	case didclient.P256KeyType:
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

	newDidDoc, err := c.vdrRegistry.Create(
		"peer",
		vdr.WithServices(did.Service{
			ServiceEndpoint: config.Endpoint(),
			RoutingKeys:     config.Keys(),
		}),
	)
	if err != nil {
		logutil.LogError(logger, CommandName, CreatePeerDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	didSvc, ok := did.LookupService(newDidDoc, didCommServiceType)
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

	bytes, err := newDidDoc.JSONBytes()
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

// SaveDID saves the DID to the DID store.
func (c *Command) SaveDID(_ io.Writer, req io.Reader) command.Error {
	didDataToStore := sdscomm.SaveDIDDocToSDSRequest{}

	err := json.NewDecoder(req).Decode(&didDataToStore)
	if err != nil {
		logutil.LogError(logger, CommandName, saveDIDCommandMethod,
			fmt.Sprintf("%s: %s", errDecodeDIDDocDataErrMsg, err.Error()))

		return command.NewValidationError(InvalidRequestErrorCode,
			fmt.Errorf("%s: %w", errDecodeDIDDocDataErrMsg, err))
	}

	return c.saveDID(&didDataToStore)
}

func (c *Command) saveDID(didDataToStore *sdscomm.SaveDIDDocToSDSRequest) command.Error {
	err := c.sdsComm.StoreDIDDocument(didDataToStore)
	if err != nil {
		logutil.LogError(logger, CommandName, saveDIDCommandMethod,
			fmt.Sprintf("%s: %s", errStoreDIDDocErrMsg, err.Error()))

		return command.NewValidationError(InvalidRequestErrorCode,
			fmt.Errorf("%s: %w", errStoreDIDDocErrMsg, err))
	}

	return nil
}
