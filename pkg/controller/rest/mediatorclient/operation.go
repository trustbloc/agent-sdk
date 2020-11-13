/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package mediatorclient provides REST operations for mediator client command.
package mediatorclient

import (
	"fmt"
	"net/http"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/mediatorclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest"
)

// constants for endpoints of mediator client.
const (
	OperationID = "/mediatorclient"
	ConnectPath = OperationID + "/connect"
)

// Operation is controller REST service controller for mediator Client.
type Operation struct {
	command  *mediatorclient.Command
	handlers []rest.Handler
}

// New returns new mediator client rest instance.
func New(ctx mediatorclient.Provider) (*Operation, error) {
	client, err := mediatorclient.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mediator-client command: %w", err)
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
		cmdutil.NewHTTPHandler(ConnectPath, http.MethodPost, c.Connect),
	}
}

// Connect swagger:route POST /mediatorclient/connect mediatorclient connectionRequest
//
// Connects to mediator.
//
// Responses:
//    default: genericError
//    200: connectionResponse
func (c *Operation) Connect(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.Connect, rw, req.Body)
}
