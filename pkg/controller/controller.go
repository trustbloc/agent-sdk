/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package controller

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	credentialclientcmd "github.com/trustbloc/agent-sdk/pkg/controller/command/credentialclient"
	didclientcmd "github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
	presentationclientcmd "github.com/trustbloc/agent-sdk/pkg/controller/command/presentationclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
)

type allOpts struct {
	blocDomain   string
	sdsServerURL string
}

// Opt represents a controller option.
type Opt func(opts *allOpts)

// WithBlocDomain is an option allowing for the trustbloc domain to be set.
func WithBlocDomain(blocDomain string) Opt {
	return func(opts *allOpts) {
		opts.blocDomain = blocDomain
	}
}

// WithSDSServerURL is an option allowing for the SDS server URL to be set.
func WithSDSServerURL(sdsServerURL string) Opt {
	return func(opts *allOpts) {
		opts.sdsServerURL = sdsServerURL
	}
}

// GetCommandHandlers returns all command handlers provided by controller.
func GetCommandHandlers(ctx *context.Provider, opts ...Opt) ([]command.Handler, error) {
	cmdOpts := &allOpts{}
	// Apply options
	for _, opt := range opts {
		opt(cmdOpts)
	}

	sdsComm := sdscomm.New(cmdOpts.sdsServerURL)

	// did client command operation
	didClientCmd, err := didclientcmd.New(cmdOpts.blocDomain, sdsComm, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DID client: %w", err)
	}

	// credential client command operation
	credentialClientCmd := credentialclientcmd.New(sdsComm)

	// presentation client command operation
	presentationClientCmd := presentationclientcmd.New(sdsComm)

	var allHandlers []command.Handler
	allHandlers = append(allHandlers, didClientCmd.GetHandlers()...)
	allHandlers = append(allHandlers, credentialClientCmd.GetHandlers()...)
	allHandlers = append(allHandlers, presentationClientCmd.GetHandlers()...)

	return allHandlers, nil
}
