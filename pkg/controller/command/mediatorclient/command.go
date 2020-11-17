/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package mediatorclient provides client features for aries mediator commands.
package mediatorclient

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
	ariescmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"
	mediatorcmd "github.com/hyperledger/aries-framework-go/pkg/controller/command/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexchangeSvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
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
	// ReconnectAll command name.
	ReconnectAll = "ReconnectAll"
	// CreateInvitation command name.
	CreateInvitation = "CreateInvitation"
)

const (
	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + command.MediatorClient)

	// ConnectMediatorError is typically a code for mediator connect errors.
	ConnectMediatorError
	// ReconnectAllError is typically a code for mediator reconnectAll errors.
	ReconnectAllError
	// CreateInvitationError is typically a code for mediator create invitation command errors.
	CreateInvitationError

	// errors.
	errInvalidConnectionRequest = "invitation missing in connection request"
	errNoConnectionFound        = "no connection found to create invitation"

	// log constants.
	successString = "success"

	// messaging & notifications.
	stateCompleteTopic = "state-complete-topic"

	// timeout constants.
	didExchangeTimeOut = 20 * time.Second
)

// Provider describes dependencies for this command.
type Provider interface {
	Service(id string) (interface{}, error)
	KMS() kms.KeyManager
	ServiceEndpoint() string
	StorageProvider() storage.Provider
	ProtocolStateStorageProvider() storage.Provider
}

// Command is controller command for mediator client.
type Command struct {
	didExchange    *didexchange.Client
	outOfBand      *outofband.Client
	mediator       *mediator.Client
	didExchTimeout time.Duration
	msgHandler     ariescmd.MessageHandler
	registry       *command.Registry
}

// New returns new mediator client controller command instance.
func New(p Provider, msgHandler ariescmd.MessageHandler) (*Command, error) {
	mediatorClient, err := mediator.New(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create mediator client : %w", err)
	}

	didExchangeClient, err := didexchange.New(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create did-exchange client : %w", err)
	}

	outOfBandClient, err := outofband.New(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create out-of-band client : %w", err)
	}

	return &Command{
		didExchange:    didExchangeClient,
		outOfBand:      outOfBandClient,
		mediator:       mediatorClient,
		didExchTimeout: didExchangeTimeOut,
		msgHandler:     msgHandler,
	}, nil
}

// SetCommandRegistry sets commands registry to this command hanler which can be used to call external
// commands.
func (c *Command) SetCommandRegistry(registry *command.Registry) {
	c.registry = registry
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []command.Handler {
	return []command.Handler{
		cmdutil.NewCommandHandler(CommandName, Connect, c.Connect),
		cmdutil.NewCommandHandler(CommandName, ReconnectAll, c.ReconnectAll),
		cmdutil.NewCommandHandler(CommandName, CreateInvitation, c.CreateInvitation),
	}
}

// Connect connects agent to given router endpoint.
func (c *Command) Connect(rw io.Writer, req io.Reader) command.Error { //nolint:funlen, gocyclo
	var request ConnectionRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	if request.Invitation == nil {
		logutil.LogError(logger, CommandName, Connect, errInvalidConnectionRequest)

		return command.NewValidationError(InvalidRequestErrorCode, fmt.Errorf(errInvalidConnectionRequest))
	}

	statusCh := make(chan service.StateMsg)

	err = c.didExchange.RegisterMsgEvent(statusCh)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewExecuteError(ConnectMediatorError, err)
	}

	var notificationCh chan msghandler.NotificationPayload

	if request.StateCompleteMessageType != "" {
		notificationCh = make(chan msghandler.NotificationPayload)

		err = c.msgHandler.Register(msghandler.NewMessageService(stateCompleteTopic, request.StateCompleteMessageType,
			nil, msghandler.NewNotifier(notificationCh)))
		if err != nil {
			logutil.LogError(logger, CommandName, Connect, err.Error())

			return command.NewExecuteError(ConnectMediatorError, err)
		}

		defer func() {
			e := c.msgHandler.Unregister(stateCompleteTopic)
			if e != nil {
				logger.Warnf("Failed to unregister state completion notifier: %w", e)
			}
		}()
	}

	connID, err := c.outOfBand.AcceptInvitation(request.Invitation, request.MyLabel)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewExecuteError(ConnectMediatorError, err)
	}

	err = c.waitForStateCompleted(statusCh, connID)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewExecuteError(ConnectMediatorError, err)
	}

	err = c.waitForStateCompletedNotification(notificationCh)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewExecuteError(ConnectMediatorError, err)
	}

	err = c.mediator.Register(connID)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewExecuteError(ConnectMediatorError, err)
	}

	config, err := c.mediator.GetConfig(connID)
	if err != nil {
		logutil.LogError(logger, CommandName, Connect, err.Error())

		return command.NewExecuteError(ConnectMediatorError, err)
	}

	command.WriteNillableResponse(rw, &ConnectionResponse{
		ConnectionID:   connID,
		RoutingKeys:    config.Keys(),
		RouterEndpoint: config.Endpoint(),
	}, logger)

	logutil.LogDebug(logger, CommandName, Connect, successString)

	return nil
}

// ReconnectAll performs reconnection on all available mediator connections.
// This command is useful in re-establishing lost connections (ex: lost websocket connection).
func (c *Command) ReconnectAll(rw io.Writer, req io.Reader) command.Error {
	connections, err := c.mediator.GetConnections()
	if err != nil {
		logutil.LogError(logger, CommandName, ReconnectAll, err.Error())

		return command.NewExecuteError(ReconnectAllError, err)
	}

	for _, connection := range connections {
		err := c.registry.Execute(mediatorcmd.CommandName,
			mediatorcmd.ReconnectCommandMethod,
			&mediatorcmd.RegisterRoute{ConnectionID: connection}, nil)
		if err != nil {
			logutil.LogError(logger, CommandName, ReconnectAll, err.Error())

			return command.NewExecuteError(ReconnectAllError, err)
		}
	}

	command.WriteNillableResponse(rw, nil, logger)

	logutil.LogDebug(logger, CommandName, ReconnectAll, successString)

	return nil
}

// CreateInvitation creates out-of-band invitation from one of the mediator connections.
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

	invitation, err := c.outOfBand.CreateInvitation(request.Protocols,
		outofband.WithGoal(request.Goal, request.GoalCode),
		outofband.WithLabel(request.Label),
		outofband.WithServices(request.Service...),
		outofband.WithRouterConnections(connections[rand.Intn(len(connections))])) //nolint: gosec
	if err != nil {
		logutil.LogError(logger, CommandName, CreateInvitation, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	command.WriteNillableResponse(rw, &CreateInvitationResponse{Invitation: invitation}, logger)

	logutil.LogDebug(logger, CommandName, CreateInvitation, successString)

	return nil
}

func (c *Command) waitForStateCompleted(didStateMsgs chan service.StateMsg, connID string) error {
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

func (c *Command) waitForStateCompletedNotification(notificationCh chan msghandler.NotificationPayload) error {
	if notificationCh == nil {
		return nil
	}

	select {
	case <-notificationCh:
		// TODO correlate connection ID
	case <-time.After(c.didExchTimeout):
		return fmt.Errorf("timeout waiting for state completed message from mediator")
	}

	return nil
}
