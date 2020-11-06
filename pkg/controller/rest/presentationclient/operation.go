/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package presentationclient provides REST operations.
package presentationclient

import (
	"net/http"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/presentationclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest"
)

// constants for endpoints of presentation client.
const (
	OperationID          = "/presentationclient"
	SavePresentationPath = OperationID + "/save-presentation"
)

// Operation is controller REST service controller for presentation client.
type Operation struct {
	command  *presentationclient.Command
	handlers []rest.Handler
}

// New returns new presentation rest instance.
func New(sdsComm *sdscomm.SDSComm) *Operation {
	o := &Operation{command: presentationclient.New(sdsComm)}
	o.registerHandler()

	return o
}

// GetRESTHandlers get all controller API handler available for this protocol service.
func (c *Operation) GetRESTHandlers() []rest.Handler {
	return c.handlers
}

// registerHandler register handlers to be exposed from this protocol service as REST API endpoints.
func (c *Operation) registerHandler() {
	// Add more protocol endpoints here to expose them as controller API endpoints
	c.handlers = []rest.Handler{
		cmdutil.NewHTTPHandler(SavePresentationPath, http.MethodPost, c.SavePresentation),
	}
}

// SavePresentation received in the request.
func (c *Operation) SavePresentation(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.SavePresentation, rw, req.Body)
}
