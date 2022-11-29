/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package didclient provides REST operations.
package didclient

import (
	"fmt"
	"net/http"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest"
)

// constants for endpoints of DIDClient.
const (
	OperationID                 = "/didclient"
	CreateOrbDIDPath            = OperationID + "/create-orb-did"
	CreatePeerDIDPath           = OperationID + "/create-peer-did"
	ResolveOrbDIDPath           = OperationID + "/resolve-orb-did"
	ResolveWebDIDFromOrbDIDPath = OperationID + "/resolve-web-did-from-orb-did"
	VerifyWebDIDFromOrbDIDPath  = OperationID + "/verify-web-did-from-orb-did"
)

// Operation is controller REST service controller for DID Client.
type Operation struct {
	command  *didclient.Command
	handlers []rest.Handler
}

// New returns new DID client rest instance.
func New(ctx didclient.ProviderWithMediator, domain, didAnchorOrigin, token string,
	unanchoredDIDMaxLifeTime int,
) (*Operation, error) {
	client, err := didclient.NewWithMediator(domain, didAnchorOrigin, token, unanchoredDIDMaxLifeTime, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize did-client command: %w", err)
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
		cmdutil.NewHTTPHandler(CreateOrbDIDPath, http.MethodPost, c.CreateOrbDID),
		cmdutil.NewHTTPHandler(CreatePeerDIDPath, http.MethodPost, c.CreatePeerDID),
		cmdutil.NewHTTPHandler(ResolveOrbDIDPath, http.MethodPost, c.ResolveOrbDID),
		cmdutil.NewHTTPHandler(ResolveWebDIDFromOrbDIDPath, http.MethodPost, c.ResolveWebDIDFromOrbDID),
		cmdutil.NewHTTPHandler(VerifyWebDIDFromOrbDIDPath, http.MethodPost, c.VerifyWebDIDFromOrbDID),
	}
}

// CreateOrbDID swagger:route POST /didclient/create-orb-did didclient createOrbDID
//
// Creates a new orb DID.
//
// Responses:
//
//	default: genericError
//	200: createDIDResp
func (c *Operation) CreateOrbDID(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.CreateOrbDID, rw, req.Body)
}

// ResolveOrbDID swagger:route POST /didclient/resolve-orb-did didclient resolveOrbDID
//
// Resolve orb DID.
//
// Responses:
//
//	default: genericError
//	200: resolveDIDResp
func (c *Operation) ResolveOrbDID(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.ResolveOrbDID, rw, req.Body)
}

// ResolveWebDIDFromOrbDID swagger:route POST /didclient/resolve-web-did-from-orb-did didclient resolveWebDIDFromOrbDID
//
// Resolve web DID from orb DID.
//
// Responses:
//
//	default: genericError
//	200: resolveDIDResp
func (c *Operation) ResolveWebDIDFromOrbDID(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.ResolveWebDIDFromOrbDID, rw, req.Body)
}

// VerifyWebDIDFromOrbDID swagger:route POST /didclient/verify-web-did-from-orb-did didclient verifyWebDIDFromOrbDID
//
// Verify web DID from orb DID.
//
// Responses:
//
//	default: genericError
func (c *Operation) VerifyWebDIDFromOrbDID(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.VerifyWebDIDFromOrbDID, rw, req.Body)
}

// CreatePeerDID swagger:route POST /didclient/create-peer-did didclient createPeerDID
//
// Creates a new peer DID.
//
// Responses:
//
//	default: genericError
//	200: createDIDResp
func (c *Operation) CreatePeerDID(rw http.ResponseWriter, req *http.Request) {
	rest.Execute(c.command.CreatePeerDID, rw, req.Body)
}
