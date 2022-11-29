/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package blindedrouting provides REST operations for blinded routing command.
package blindedrouting

import (
	"fmt"
	"net/http"

	ariescmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/blindedrouting"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest"
)

// constants for endpoints of blinded routing.
const (
	OperationID              = "/blindedrouting"
	SendDIDDocRequestPath    = OperationID + "/send-diddoc-request"
	SendRegisterRouteRequest = OperationID + "/send-router-registration"
)

// Operation is controller REST service controller for blinded routing.
type Operation struct {
	command  *blindedrouting.Command
	handlers []rest.Handler
}

// New returns new blinded routing rest instance.
func New(ctx blindedrouting.Provider, msgHandler ariescmd.MessageHandler,
	notifier ariescmd.Notifier,
) (*Operation, error) {
	client, err := blindedrouting.New(ctx, msgHandler, notifier)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize blinded routing command: %w", err)
	}

	o := &Operation{command: client}
	o.registerHandler()

	return o, nil
}

// GetRESTHandlers get all controller API handler available for this protocol service.
func (c *Operation) GetRESTHandlers() []rest.Handler {
	return c.handlers
}

// registerHandler register handlers to be exposed from this protocol service as REST API endpoints.
func (c *Operation) registerHandler() {
	// Add more protocol endpoints here to expose them as controller API endpoints
	c.handlers = []rest.Handler{
		cmdutil.NewHTTPHandler(SendDIDDocRequestPath, http.MethodPost, c.SendDIDDocRequest),
		cmdutil.NewHTTPHandler(SendRegisterRouteRequest, http.MethodPost, c.SendRegisterRouteRequest),
	}
}

// SendDIDDocRequest swagger:route POST /blindedrouting/send-diddoc-request blindedrouting didDocRequest
//
// Sends DID doc request over a connection.
//
// Responses:
//
//	default: genericError
//	200: didDocResponse
func (c *Operation) SendDIDDocRequest(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.SendDIDDocRequest, rw, req.Body)
}

// SendRegisterRouteRequest Sends register route request as a response to reply from send DID doc request.
//
// swagger:route POST /blindedrouting/send-router-registration blindedrouting registerRoute
//
// Sends register route request as a response to reply from send DID doc request.
//
// Responses:
//
//	default: genericError
//	200: registerRouteResponse
func (c *Operation) SendRegisterRouteRequest(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.SendRegisterRouteRequest, rw, req.Body)
}
