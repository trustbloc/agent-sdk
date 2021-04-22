/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package controller provides command handlers.
package controller

import (
	"fmt"

	ariescmd "github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/hyperledger/aries-framework-go/pkg/controller/webnotifier"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/blindedrouting"
	didclientcmd "github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
	mediatorclientcmd "github.com/trustbloc/agent-sdk/pkg/controller/command/mediatorclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/store"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest"
	blindedroutingrest "github.com/trustbloc/agent-sdk/pkg/controller/rest/blindedrouting"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest/didclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/rest/mediatorclient"
)

const wsPath = "/ws"

type allOpts struct {
	blocDomain      string
	didAnchorOrigin string
	msgHandler      ariescmd.MessageHandler
	notifier        ariescmd.Notifier
	webhookURLs     []string
}

// Opt represents a controller option.
type Opt func(opts *allOpts)

// WithBlocDomain is an option allowing for the trustbloc domain to be set.
func WithBlocDomain(blocDomain string) Opt {
	return func(opts *allOpts) {
		opts.blocDomain = blocDomain
	}
}

// WithDidAnchorOrigin is an option allowing for the did anchor origin.
func WithDidAnchorOrigin(origin string) Opt {
	return func(opts *allOpts) {
		opts.didAnchorOrigin = origin
	}
}

// WithMessageHandler is an option allowing for the message handler to be set.
func WithMessageHandler(handler ariescmd.MessageHandler) Opt {
	return func(opts *allOpts) {
		opts.msgHandler = handler
	}
}

// WithWebhookURLs is an option for setting up a webhook dispatcher which will notify clients of events.
func WithWebhookURLs(webhookURLs ...string) Opt {
	return func(opts *allOpts) {
		opts.webhookURLs = webhookURLs
	}
}

// WithNotifier is an option for setting up a notifier which will notify clients of events.
func WithNotifier(notifier ariescmd.Notifier) Opt {
	return func(opts *allOpts) {
		opts.notifier = notifier
	}
}

// GetCommandHandlers returns all command handlers provided by controller.
func GetCommandHandlers(ctx *context.Provider, opts ...Opt) ([]command.Handler, error) { //nolint:interfacer
	cmdOpts := &allOpts{}
	// Apply options
	for _, opt := range opts {
		opt(cmdOpts)
	}

	notifier := cmdOpts.notifier
	if notifier == nil {
		notifier = webnotifier.New(wsPath, cmdOpts.webhookURLs)
	}

	// did client command operation.
	didClientCmd, err := didclientcmd.New(cmdOpts.blocDomain, cmdOpts.didAnchorOrigin, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DID client: %w", err)
	}

	// mediator client command operation,
	mediatorClientCmd, err := mediatorclientcmd.New(ctx, cmdOpts.msgHandler, notifier)
	if err != nil {
		return nil, err
	}

	// blindedRoutingCmd command operation.
	blindedRoutingCmd, err := blindedrouting.New(ctx, cmdOpts.msgHandler, notifier)
	if err != nil {
		return nil, err
	}

	storeCmd, err := store.New(ctx)
	if err != nil {
		return nil, err
	}

	// creat handlers for all command operations.
	var allHandlers []command.Handler
	allHandlers = append(allHandlers, didClientCmd.GetHandlers()...)
	allHandlers = append(allHandlers, mediatorClientCmd.GetHandlers()...)
	allHandlers = append(allHandlers, blindedRoutingCmd.GetHandlers()...)
	allHandlers = append(allHandlers, storeCmd.GetHandlers()...)

	return allHandlers, nil
}

// GetRESTHandlers returns all REST handlers provided by controller.
func GetRESTHandlers(ctx *context.Provider, opts ...Opt) ([]rest.Handler, error) { //nolint:interfacer
	restOpts := &allOpts{}
	// Apply options
	for _, opt := range opts {
		opt(restOpts)
	}

	notifier := restOpts.notifier
	if notifier == nil {
		notifier = webnotifier.New(wsPath, restOpts.webhookURLs)
	}

	// DID Client REST operation.
	didClientOp, err := didclient.New(ctx, restOpts.blocDomain, restOpts.didAnchorOrigin)
	if err != nil {
		return nil, err
	}

	// mediator client REST operation.
	mediatorClientOp, err := mediatorclient.New(ctx, restOpts.msgHandler, notifier)
	if err != nil {
		return nil, err
	}

	// blinded routing REST operation.
	blindedRoutingOp, err := blindedroutingrest.New(ctx, restOpts.msgHandler, notifier)
	if err != nil {
		return nil, err
	}

	// creat handlers from all REST operations.
	var allHandlers []rest.Handler
	allHandlers = append(allHandlers, didClientOp.GetRESTHandlers()...)
	allHandlers = append(allHandlers, mediatorClientOp.GetRESTHandlers()...)
	allHandlers = append(allHandlers, blindedRoutingOp.GetRESTHandlers()...)

	return allHandlers, nil
}
