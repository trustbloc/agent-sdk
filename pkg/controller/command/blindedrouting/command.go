/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package blindedrouting provides blinded routing features for agents.
package blindedrouting

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/client/messaging"
	ariescmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/spi/storage"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/logutil"
)

var logger = log.New("agent-sdk-mediatorclient")

const (
	// CommandName package command name.
	CommandName = "blindedrouting"
	// SendDIDDocRequest command name.
	SendDIDDocRequest = "SendDIDDocRequest"
	// SendRegisterRouteRequest command name.
	SendRegisterRouteRequest = "SendRegisterRouteRequest"
)

const (
	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = ariescmd.Code(iota + command.MediatorClient)

	// SendDIDDocRequestError is typically a code for send did doc request command errors.
	SendDIDDocRequestError
	// SendRegisterRouteRequestError is typically a code for send register route request command errors.
	SendRegisterRouteRequestError

	// errors.
	errInvalidConnectionID = "invalid connection ID"
	errInvalidMessageID    = "invalid message ID"

	// log constants.
	successString = "success"

	// timeout constants.
	sendMsgTimeOut = 20 * time.Second

	// message types.
	didDocRequestMsgType         = "https://trustbloc.dev/blinded-routing/1.0/diddoc-req"
	didDocResponseMsgType        = "https://trustbloc.dev/blinded-routing/1.0/diddoc-resp"
	registerRouteRequestMsgType  = "https://trustbloc.dev/blinded-routing/1.0/register-route-req"
	registerRouteResponseMsgType = "https://trustbloc.dev/blinded-routing/1.0/register-route-resp"
)

// Provider describes dependencies for this command.
type Provider interface {
	VDRegistry() vdr.Registry
	Messenger() service.Messenger
	ProtocolStateStorageProvider() storage.Provider
	StorageProvider() storage.Provider
	KMS() kms.KeyManager
}

// Command is controller command for blinded routing.
type Command struct {
	messenger *messaging.Client
}

// New returns new blinded routing controller command instance.
func New(p Provider, msgHandler ariescmd.MessageHandler, notifier ariescmd.Notifier) (*Command, error) {
	messengerClient, err := messaging.New(p, msgHandler, notifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create messenger client : %w", err)
	}

	return &Command{
		messenger: messengerClient,
	}, nil
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []ariescmd.Handler {
	return []ariescmd.Handler{
		cmdutil.NewCommandHandler(CommandName, SendDIDDocRequest, c.SendDIDDocRequest),
		cmdutil.NewCommandHandler(CommandName, SendRegisterRouteRequest, c.SendRegisterRouteRequest),
	}
}

// SendDIDDocRequest sends DID doc request over a connection.
func (c *Command) SendDIDDocRequest(rw io.Writer, req io.Reader) ariescmd.Error {
	var request DIDDocRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, SendDIDDocRequest, err.Error())

		return ariescmd.NewValidationError(InvalidRequestErrorCode, err)
	}

	if request.ConnectionID == "" {
		logutil.LogError(logger, CommandName, SendDIDDocRequest, errInvalidConnectionID)

		return ariescmd.NewValidationError(SendDIDDocRequestError, fmt.Errorf(errInvalidConnectionID))
	}

	msgStr := fmt.Sprintf(`{"@id":%q,"@type":%q}`, uuid.New().String(), didDocRequestMsgType)

	ctx, cancel := context.WithTimeout(context.Background(), sendMsgTimeOut)
	defer cancel()

	resMsg, err := c.messenger.Send(json.RawMessage([]byte(msgStr)),
		messaging.SendByConnectionID(request.ConnectionID),
		messaging.WaitForResponse(ctx, didDocResponseMsgType))
	if err != nil {
		logutil.LogError(logger, CommandName, SendDIDDocRequest, err.Error())

		return ariescmd.NewExecuteError(SendDIDDocRequestError, err)
	}

	command.WriteNillableResponse(rw, &DIDDocResponse{resMsg}, logger)

	logutil.LogDebug(logger, CommandName, SendDIDDocRequest, successString)

	return nil
}

// SendRegisterRouteRequest sends register route request as a response to reply from send DID doc request.
func (c *Command) SendRegisterRouteRequest(rw io.Writer, req io.Reader) ariescmd.Error {
	var request RegisterRouteRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, SendRegisterRouteRequest, err.Error())

		return ariescmd.NewValidationError(InvalidRequestErrorCode, err)
	}

	if request.MessageID == "" {
		logutil.LogError(logger, CommandName, SendRegisterRouteRequest, errInvalidMessageID)

		return ariescmd.NewValidationError(SendRegisterRouteRequestError, fmt.Errorf(errInvalidMessageID))
	}

	msgBytes, err := json.Marshal(map[string]interface{}{
		"@id":   uuid.New().String(),
		"@type": registerRouteRequestMsgType,
		"data": map[string]interface{}{
			"didDoc": request.DIDDocument,
		},
	})
	if err != nil {
		logutil.LogError(logger, CommandName, SendRegisterRouteRequest, err.Error())

		return ariescmd.NewValidationError(SendRegisterRouteRequestError, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sendMsgTimeOut)
	defer cancel()

	res, err := c.messenger.Reply(ctx, msgBytes, request.MessageID, true, registerRouteResponseMsgType)
	if err != nil {
		logutil.LogError(logger, CommandName, SendRegisterRouteRequest, err.Error())

		return ariescmd.NewExecuteError(SendRegisterRouteRequestError, err)
	}

	command.WriteNillableResponse(rw, &RegisterRouteResponse{res}, logger)

	logutil.LogDebug(logger, CommandName, SendRegisterRouteRequest, successString)

	return nil
}
