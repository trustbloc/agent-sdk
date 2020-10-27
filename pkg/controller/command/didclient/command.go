/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package didclient

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	mediatorservice "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/internal/logutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
	"github.com/trustbloc/edge-core/pkg/log"
	didclient "github.com/trustbloc/trustbloc-did-method/pkg/did"
)

var logger = log.New("agent-sdk-didclient")

const (
	// command name.
	commandName = "didclient"
	// command methods.
	createBlocDIDCommandMethod = "CreateBlocDID"
	createPeerDIDCommandMethod = "CreatePeerDID"
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

	// errors
	errDecodeDIDDocDataErrMsg    = "failure while decoding DID data"
	errStoreDIDDocErrMsg         = "failure while storing DID document in SDS"
	errInvalidRouterConnectionID = "invalid router connection ID"
	errMissingDIDCommServiceType = "did document missing '%s' service type"
	errFailedToRegisterDIDRecKey = "failed to register did doc recipient key : %w"
)

type provider interface {
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
func New(domain string, sdsComm *sdscomm.SDSComm, p provider) (*Command, error) {
	client := didclient.New()

	mediator, err := mediator.New(p)
	if err != nil {
		return nil, err
	}

	s, err := p.Service(mediatorservice.Coordination)
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
		mediatorClient: mediator,
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
		cmdutil.NewCommandHandler(commandName, createBlocDIDCommandMethod, c.CreateTrustBlocDID),
		cmdutil.NewCommandHandler(commandName, createPeerDIDCommandMethod, c.CreatePeerDID),
		cmdutil.NewCommandHandler(commandName, saveDIDCommandMethod, c.SaveDID),
	}
}

// CreateTrustBlocDID creates a new trust bloc DID.
func (c *Command) CreateTrustBlocDID(rw io.Writer, req io.Reader) command.Error {
	var request CreateBlocDIDRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, commandName, createBlocDIDCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	var opts []didclient.CreateDIDOption

	for _, v := range request.PublicKeys {
		value, decodeErr := base64.RawURLEncoding.DecodeString(v.Value)
		if decodeErr != nil {
			logutil.LogError(logger, commandName, createBlocDIDCommandMethod, decodeErr.Error())

			return command.NewExecuteError(CreateDIDErrorCode, decodeErr)
		}

		opts = append(opts, didclient.WithPublicKey(&didclient.PublicKey{
			ID: v.ID, Type: v.Type, Encoding: v.Encoding,
			KeyType: v.KeyType, Purpose: v.Purpose, Recovery: v.Recovery, Update: v.Update, Value: value,
		}))
	}

	didDoc, err := c.didBlocClient.CreateDID(c.domain, opts...)
	if err != nil {
		logutil.LogError(logger, commandName, createBlocDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	bytes, err := didDoc.JSONBytes()
	if err != nil {
		logutil.LogError(logger, commandName, createBlocDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	command.WriteNillableResponse(rw, &CreateDIDResponse{
		DID: bytes,
	}, logger)

	logutil.LogDebug(logger, commandName, createBlocDIDCommandMethod, successString)

	return nil
}

// CreatePeerDID creates a new peer DID.
func (c *Command) CreatePeerDID(rw io.Writer, req io.Reader) command.Error {
	var request CreatePeerDIDRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, commandName, createPeerDIDCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	if request.RouterConnectionID == "" {
		logutil.LogError(logger, commandName, createPeerDIDCommandMethod, errInvalidRouterConnectionID)

		return command.NewValidationError(InvalidRequestErrorCode, fmt.Errorf(errInvalidRouterConnectionID))
	}

	config, err := c.mediatorClient.GetConfig(request.RouterConnectionID)
	if err != nil {
		logutil.LogError(logger, commandName, createPeerDIDCommandMethod, err.Error())

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
		logutil.LogError(logger, commandName, createPeerDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	didSvc, ok := did.LookupService(newDidDoc, didCommServiceType)
	if !ok {
		logutil.LogError(logger, commandName, createPeerDIDCommandMethod, errMissingDIDCommServiceType)

		return command.NewExecuteError(CreateDIDErrorCode, fmt.Errorf(errMissingDIDCommServiceType, didCommServiceType))
	}

	for _, val := range didSvc.RecipientKeys {
		err = mediatorservice.AddKeyToRouter(c.mediatorSvc, request.RouterConnectionID, val)

		if err != nil {
			logutil.LogError(logger, commandName, createPeerDIDCommandMethod, err.Error())

			return command.NewExecuteError(CreateDIDErrorCode, fmt.Errorf(errFailedToRegisterDIDRecKey, err))
		}
	}

	bytes, err := newDidDoc.JSONBytes()
	if err != nil {
		logutil.LogError(logger, commandName, createBlocDIDCommandMethod, err.Error())

		return command.NewExecuteError(CreateDIDErrorCode, err)
	}

	command.WriteNillableResponse(rw, &CreateDIDResponse{
		DID: bytes,
	}, logger)

	logutil.LogDebug(logger, commandName, createBlocDIDCommandMethod, successString)

	return nil
}

// SaveDID saves the DID to the DID store.
func (c *Command) SaveDID(_ io.Writer, req io.Reader) command.Error {
	didDataToStore := sdscomm.SaveDIDDocToSDSRequest{}

	err := json.NewDecoder(req).Decode(&didDataToStore)
	if err != nil {
		logutil.LogError(logger, commandName, saveDIDCommandMethod,
			fmt.Sprintf("%s: %s", errDecodeDIDDocDataErrMsg, err.Error()))

		return command.NewValidationError(InvalidRequestErrorCode,
			fmt.Errorf("%s: %w", errDecodeDIDDocDataErrMsg, err))
	}

	return c.saveDID(&didDataToStore)
}

func (c *Command) saveDID(didDataToStore *sdscomm.SaveDIDDocToSDSRequest) command.Error {
	err := c.sdsComm.StoreDIDDocument(didDataToStore)
	if err != nil {
		logutil.LogError(logger, commandName, saveDIDCommandMethod,
			fmt.Sprintf("%s: %s", errStoreDIDDocErrMsg, err.Error()))

		return command.NewValidationError(InvalidRequestErrorCode,
			fmt.Errorf("%s: %w", errStoreDIDDocErrMsg, err))
	}

	return nil
}
