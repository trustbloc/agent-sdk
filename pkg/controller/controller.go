/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package controller provides command handlers.
package controller

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/framework/context"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	didclientcmd "github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest/didclient"
)

type allOpts struct {
	blocDomain string
}

// Opt represents a controller option.
type Opt func(opts *allOpts)

// WithBlocDomain is an option allowing for the trustbloc domain to be set.
func WithBlocDomain(blocDomain string) Opt {
	return func(opts *allOpts) {
		opts.blocDomain = blocDomain
	}
}

// GetCommandHandlers returns all command handlers provided by controller.
func GetCommandHandlers(ctx *context.Provider, opts ...Opt) ([]command.Handler, error) { //nolint: interfacer
	cmdOpts := &allOpts{}
	// Apply options
	for _, opt := range opts {
		opt(cmdOpts)
	}

	// did client command operation
	didClientCmd, err := didclientcmd.New(cmdOpts.blocDomain, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DID client: %w", err)
	}

	var allHandlers []command.Handler
	allHandlers = append(allHandlers, didClientCmd.GetHandlers()...)

	return allHandlers, nil
}

// GetRESTHandlers returns all REST handlers provided by controller.
func GetRESTHandlers(ctx *context.Provider, opts ...Opt) ([]rest.Handler, error) { //nolint: interfacer
	restOpts := &allOpts{}
	// Apply options
	for _, opt := range opts {
		opt(restOpts)
	}

	// DID Client REST operation
	didClientOp, err := didclient.New(ctx, restOpts.blocDomain)
	if err != nil {
		return nil, err
	}

	// creat handlers from all operations
	var allHandlers []rest.Handler
	allHandlers = append(allHandlers, didClientOp.GetRESTHandlers()...)

	return allHandlers, nil
}
