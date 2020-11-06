/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialclient provides REST operations.
package credentialclient

import (
	"net/http"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/credentialclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest"
)

// constants for endpoints of credential client.
const (
	OperationID        = "/credentialclient"
	SaveCredentialPath = OperationID + "/save-credential"
)

// Operation is controller REST service controller for credential client.
type Operation struct {
	command  *credentialclient.Command
	handlers []rest.Handler
}

// New returns new credential rest instance.
func New(sdsComm *sdscomm.SDSComm) *Operation {
	o := &Operation{command: credentialclient.New(sdsComm)}
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
		cmdutil.NewHTTPHandler(SaveCredentialPath, http.MethodPost, c.SaveCredential),
	}
}

// SaveCredential received in the request.
func (c *Operation) SaveCredential(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.SaveCredential, rw, req.Body)
}
