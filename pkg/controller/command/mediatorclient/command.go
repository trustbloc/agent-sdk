/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package mediatorclient provides client features for aries mediator commands.
package mediatorclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/client/messaging"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofbandv2"
	ariescmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexchangeSvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	oobv2 "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofbandv2"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/spi/storage"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/logutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/msghandler"
)

var logger = log.New("agent-sdk-mediatorclient")

const (
	// CommandName package command name.
	CommandName = "mediatorclient"
	// Connect command name.
	Connect = "Connect"
	// CreateInvitation command name.
	CreateInvitation = "CreateInvitation"
	// SendCreateConnectionRequest command name.
	SendCreateConnectionRequest = "SendCreateConnectionRequest"
)

const (
	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + command.MediatorClient)

	// ConnectMediatorError is typically a code for mediator connect errors.
	ConnectMediatorError
	// CreateInvitationError is typically a code for mediator create invitation command errors.
	CreateInvitationError
	// SendCreateConnectionRequestError is typically a code for mediator send create connection request command errors.
	SendCreateConnectionRequestError

	// errors.
	errInvalidConnectionRequest = "invitation missing in connection request"
	errNoConnectionFound        = "no connection found to create invitation"

	// log constants.
	successString = "success"

	// messaging & notifications.
	stateCompleteTopic = "state-complete-topic"

	// timeout constants.
	didExchangeTimeOut = 120 * time.Second
	sendMsgTimeOut     = 120 * time.Second

	// mediator connector queue buffer.
	msgEventBufferSize = 10

	// message types.
	createConnRequestMsgType  = "https://trustbloc.dev/blinded-routing/1.0/create-conn-req"
	createConnResponseMsgType = "https://trustbloc.dev/blinded-routing/1.0/create-conn-resp"
)

// Provider describes dependencies for this command.
type Provider interface {
	Service(id string) (interface{}, error)
	KMS() kms.KeyManager
	ServiceEndpoint() string
	StorageProvider() storage.Provider
	ProtocolStateStorageProvider() storage.Provider
	VDRegistry() vdr.Registry
	Messenger() service.Messenger
	KeyType() kms.KeyType
	KeyAgreementType() kms.KeyType
	MediaTypeProfiles() []string
}

// Command is controller command for mediator client.
type Command struct {
	didExchange    *didexchange.Client
	outOfBand      *outofband.Client
	outOfBandV2    *outofbandv2.Client
	mediator       *mediator.Client
	messenger      *messaging.Client
	didExchTimeout time.Duration
	msgHandler     ariescmd.MessageHandler
}

// New returns new mediator client controller command instance.
func New(p Provider, msgHandler ariescmd.MessageHandler, notifier ariescmd.Notifier) (*Command, error) {
	mediatorClient, err := mediator.New(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create mediator client : %w", err)
	}

	messengerClient, err := messaging.New(p, msgHandler, notifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create messenger client : %w", err)
	}

	didExchangeClient, err := didexchange.New(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create did-exchange client : %w", err)
	}

	outOfBandClient, err := outofband.New(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create out-of-band client : %w", err)
	}

	outOfBandClientV2, err := outofbandv2.New(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create out-of-band v2 client : %w", err)
	}

	return &Command{
		didExchange:    didExchangeClient,
		outOfBand:      outOfBandClient,
		outOfBandV2:    outOfBandClientV2,
		mediator:       mediatorClient,
		messenger:      messengerClient,
		didExchTimeout: didExchangeTimeOut,
		msgHandler:     msgHandler,
	}, nil
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []command.Handler {
	return []command.Handler{
		cmdutil.NewCommandHandler(CommandName, Connect, c.Connect),
		cmdutil.NewCommandHandler(CommandName, CreateInvitation, c.CreateInvitation),
		cmdutil.NewCommandHandler(CommandName, SendCreateConnectionRequest, c.SendCreateConnectionRequest),
	}
}

// Connect connects agent to given router endpoint.
func (c *Command) Connect(rw io.Writer, req io.Reader) command.Error {
	var (
		request ConnectionRequest
		connID  string
		isV2    bool
	)

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	if request.Invitation == nil {
		logutil.LogError(logger, CommandName, Connect, errInvalidConnectionRequest)

		return command.NewValidationError(InvalidRequestErrorCode, fmt.Errorf(errInvalidConnectionRequest))
	}

	//nolint:nestif
	if isV2, err = service.IsDIDCommV2(request.Invitation); isV2 && err == nil {
		inv := &oobv2.Invitation{}

		err = request.Invitation.Decode(inv)
		if err != nil {
			return command.NewExecuteError(ConnectMediatorError, err)
		}

		connID, err = c.outOfBandV2.AcceptInvitation(inv)
		if err != nil {
			logutil.LogError(logger, CommandName, Connect, err.Error())

			return command.NewExecuteError(ConnectMediatorError, err)
		}
	} else {
		inv := &outofband.Invitation{}
		err = request.Invitation.Decode(inv)
		if err != nil {
			return command.NewExecuteError(ConnectMediatorError, err)
		}

		connID, err = c.createOOBInvitation(inv, request.MyLabel, request.StateCompleteMessageType)
		if err != nil {
			return command.NewExecuteError(ConnectMediatorError, err)
		}
	}

	err = c.mediator.Register(connID)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewExecuteError(ConnectMediatorError, err)
	}

	command.WriteNillableResponse(rw, &ConnectionResponse{ConnectionID: connID}, logger)

	logutil.LogDebug(logger, CommandName, Connect, successString)

	return nil
}

func (c *Command) createOOBInvitation(inv *outofband.Invitation,
	myLabel, stateCompleteMessageType string) (string, error) {
	var notificationCh chan messaging.NotificationPayload

	var statusCh chan service.StateMsg

	if stateCompleteMessageType != "" { //nolint:nestif
		notificationCh = make(chan messaging.NotificationPayload)

		err := c.msgHandler.Register(msghandler.NewMessageService(stateCompleteTopic, stateCompleteMessageType,
			nil, messaging.NewNotifier(notificationCh, nil)))
		if err != nil {
			logutil.LogError(logger, CommandName, Connect, err.Error())

			return "", err
		}

		defer func() {
			e := c.msgHandler.Unregister(stateCompleteTopic)
			if e != nil {
				logger.Warnf("Failed to unregister state completion notifier: %w", e)
			}
		}()
	} else {
		statusCh = make(chan service.StateMsg, msgEventBufferSize)

		err := c.didExchange.RegisterMsgEvent(statusCh)
		if err != nil {
			logutil.LogError(logger, CommandName, Connect, err.Error())

			return "", err
		}

		defer func() {
			e := c.didExchange.UnregisterMsgEvent(statusCh)
			if e != nil {
				logger.Warnf("Failed to unregister msg event: %w", e)
			}
		}()
	}

	connID, err := c.outOfBand.AcceptInvitation(inv, myLabel)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return "", err
	}

	err = c.waitForConnect(statusCh, notificationCh, connID)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return "", err
	}

	return connID, nil
}

// CreateInvitation creates out-of-band invitation from one of the mediator connections.
//nolint:funlen
func (c *Command) CreateInvitation(rw io.Writer, req io.Reader) command.Error {
	connections, err := c.mediator.GetConnections()
	if err != nil {
		logutil.LogError(logger, CommandName, CreateInvitation, err.Error())

		return command.NewExecuteError(CreateInvitationError, err)
	}

	if len(connections) == 0 {
		logutil.LogError(logger, CommandName, CreateInvitation, errNoConnectionFound)

		return command.NewExecuteError(CreateInvitationError, fmt.Errorf(errNoConnectionFound))
	}

	var request CreateInvitationRequest

	err = json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, CreateInvitation, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	var (
		invitation   *outofband.Invitation
		invitationV2 *oobv2.Invitation
	)

	if request.From != "" {
		invitationV2, err = c.outOfBandV2.CreateInvitation(
			outofbandv2.WithAccept("didcomm/aip2;env=rfc587", "didcomm/v2"),
			outofbandv2.WithGoal(request.Goal, request.GoalCode),
			outofbandv2.WithLabel(request.Label),
			outofbandv2.WithFrom(request.From),
		)

		if err != nil {
			logutil.LogError(logger, CommandName, CreateInvitation, fmt.Sprintf("oob v2 error: %s", err.Error()))

			return command.NewValidationError(InvalidRequestErrorCode, err)
		}

		command.WriteNillableResponse(rw, &CreateInvitationResponse{InvitationV2: invitationV2}, logger)
	} else {
		invitation, err = c.outOfBand.CreateInvitation(
			request.Service,
			outofband.WithHandshakeProtocols(request.Protocols...),
			outofband.WithGoal(request.Goal, request.GoalCode),
			outofband.WithLabel(request.Label),
			outofband.WithAccept("didcomm/aip2;env=rfc19", "didcomm/aip1"),
			outofband.WithRouterConnections(connections[rand.Intn(len(connections))])) //nolint: gosec
		if err != nil {
			logutil.LogError(logger, CommandName, CreateInvitation, fmt.Sprintf("oob v1 error: %s", err.Error()))

			return command.NewValidationError(InvalidRequestErrorCode, err)
		}

		command.WriteNillableResponse(rw, &CreateInvitationResponse{Invitation: invitation}, logger)
	}

	logutil.LogDebug(logger, CommandName, CreateInvitation, fmt.Sprintf("%s for %s", successString, request.Label))

	return nil
}

// SendCreateConnectionRequest sends create connection request to mediator.
func (c *Command) SendCreateConnectionRequest(rw io.Writer, req io.Reader) command.Error {
	connections, err := c.mediator.GetConnections()
	if err != nil {
		logutil.LogError(logger, CommandName, SendCreateConnectionRequest, err.Error())

		return command.NewExecuteError(SendCreateConnectionRequestError, err)
	}

	if len(connections) == 0 {
		logutil.LogError(logger, CommandName, SendCreateConnectionRequest, errNoConnectionFound)

		return command.NewExecuteError(SendCreateConnectionRequestError, fmt.Errorf(errNoConnectionFound))
	}

	var request CreateConnectionRequest

	err = json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, SendCreateConnectionRequest, err.Error())

		return command.NewValidationError(SendCreateConnectionRequestError, err)
	}

	msgBytes, err := json.Marshal(map[string]interface{}{
		"@id":   uuid.New().String(),
		"@type": createConnRequestMsgType,
		"data": map[string]interface{}{
			"didDoc": request.DIDDocument,
		},
	})
	if err != nil {
		logutil.LogError(logger, CommandName, SendCreateConnectionRequest, err.Error())

		return command.NewValidationError(SendCreateConnectionRequestError, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sendMsgTimeOut)
	defer cancel()

	res, err := c.messenger.Send(json.RawMessage(msgBytes),
		messaging.SendByConnectionID(connections[rand.Intn(len(connections))]), //nolint: gosec
		messaging.WaitForResponse(ctx, createConnResponseMsgType))
	if err != nil {
		logutil.LogError(logger, CommandName, SendCreateConnectionRequest, err.Error())

		return command.NewExecuteError(SendCreateConnectionRequestError, err)
	}

	command.WriteNillableResponse(rw, &CreateConnectionResponse{res}, logger)

	logutil.LogDebug(logger, CommandName, SendCreateConnectionRequest, successString)

	return nil
}

//nolint: gocyclo
func (c *Command) waitForConnect(didStateMsgs chan service.StateMsg,
	notificationCh chan messaging.NotificationPayload, connID string) error {
	if notificationCh != nil {
		select {
		case <-notificationCh:
			// TODO correlate connection ID
			return nil
		case <-time.After(c.didExchTimeout):
			return fmt.Errorf("timeout waiting for state completed message from mediator")
		}
	}

	done := make(chan struct{})

	go func() {
		for msg := range didStateMsgs {
			if msg.Type != service.PostState || msg.StateID != didexchangeSvc.StateIDCompleted {
				continue
			}

			var event didexchange.Event

			switch p := msg.Properties.(type) {
			case didexchange.Event:
				event = p
			default:
				logger.Warnf("failed to cast didexchange event properties")

				continue
			}

			if event.ConnectionID() == connID {
				logger.Debugf(
					"Received connection complete event for invitationID=%s connectionID=%s",
					event.InvitationID(), event.ConnectionID())

				close(done)

				break
			}
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(c.didExchTimeout):
		return fmt.Errorf("time out waiting for did exchange state 'completed'")
	}
}
