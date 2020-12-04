/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package store provides store commands.
package store

import (
	"encoding/json"
	"io"

	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/logutil"
)

const (
	// CommandName package command name.
	CommandName = "store"
	// GetCommandMethod command method.
	GetCommandMethod = "Get"
	// PutCommandMethod command method.
	PutCommandMethod = "Put"
	// IteratorCommandMethod command method.
	IteratorCommandMethod = "Iterator"
	// DeleteCommandMethod command method.
	DeleteCommandMethod = "Delete"
	// FlushCommandMethod command method.
	FlushCommandMethod = "Flush"

	successString = "success"
)

const (
	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + command.Store)
	// GetErrorCode is typically a code for Get errors.
	GetErrorCode
	// PutErrorCode is typically a code for Put errors.
	PutErrorCode
	// IteratorErrorCode is typically a code for Iterator errors.
	IteratorErrorCode
	// DeleteErrorCode is typically a code for Delete errors.
	DeleteErrorCode
)

type batch interface {
	Flush() error
}

var logger = log.New("agent-sdk-store")

// Provider describes dependencies for the client.
type Provider interface {
	StorageProvider() storage.Provider
}

// Command is controller command for store.
type Command struct {
	store storage.Store
	batch batch
}

// New returns new store controller command instance.
func New(p Provider) (*Command, error) {
	b, ok := p.StorageProvider().(batch)
	if !ok {
		logger.Infof("store provider not supporting batch")
	}

	store, err := p.StorageProvider().OpenStore(CommandName)
	if err != nil {
		return nil, err
	}

	return &Command{store: store, batch: b}, nil
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []command.Handler {
	return []command.Handler{
		cmdutil.NewCommandHandler(CommandName, GetCommandMethod, c.Get),
		cmdutil.NewCommandHandler(CommandName, PutCommandMethod, c.Put),
		cmdutil.NewCommandHandler(CommandName, IteratorCommandMethod, c.Iterator),
		cmdutil.NewCommandHandler(CommandName, DeleteCommandMethod, c.Delete),
		cmdutil.NewCommandHandler(CommandName, FlushCommandMethod, c.Flush),
	}
}

// Get fetches the record based on key.
func (c *Command) Get(rw io.Writer, req io.Reader) command.Error {
	var request GetRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, GetCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	result, err := c.store.Get(request.Key)
	if err != nil {
		logutil.LogError(logger, CommandName, GetCommandMethod, err.Error())

		return command.NewExecuteError(GetErrorCode, err)
	}

	command.WriteNillableResponse(rw, &GetResponse{
		Result: result,
	}, logger)

	logutil.LogDebug(logger, CommandName, GetCommandMethod, successString)

	return nil
}

// Put stores the key and the record.
func (c *Command) Put(rw io.Writer, req io.Reader) command.Error {
	var request PutRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, PutCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	if err = c.store.Put(request.Key, request.Value); err != nil {
		logutil.LogError(logger, CommandName, PutCommandMethod, err.Error())

		return command.NewExecuteError(PutErrorCode, err)
	}

	command.WriteNillableResponse(rw, nil, logger)

	logutil.LogDebug(logger, CommandName, PutCommandMethod, successString)

	return nil
}

// Iterator retrieves data according to given start and end keys.
func (c *Command) Iterator(rw io.Writer, req io.Reader) command.Error {
	var request IteratorRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, IteratorCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	records := c.store.Iterator(request.StartKey, request.EndKey)
	defer records.Release()

	var values [][]byte

	for records.Next() {
		values = append(values, records.Value())
	}

	if err = records.Error(); err != nil {
		logutil.LogError(logger, CommandName, IteratorCommandMethod, err.Error())

		return command.NewExecuteError(IteratorErrorCode, err)
	}

	command.WriteNillableResponse(rw, &IteratorResponse{Results: values}, logger)

	logutil.LogDebug(logger, CommandName, IteratorCommandMethod, successString)

	return nil
}

// Delete deletes a record with a given key.
func (c *Command) Delete(rw io.Writer, req io.Reader) command.Error {
	var request DeleteRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, DeleteCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	if err = c.store.Delete(request.Key); err != nil {
		logutil.LogError(logger, CommandName, DeleteCommandMethod, err.Error())

		return command.NewExecuteError(DeleteErrorCode, err)
	}

	command.WriteNillableResponse(rw, nil, logger)

	logutil.LogDebug(logger, CommandName, DeleteCommandMethod, successString)

	return nil
}

// Flush data.
func (c *Command) Flush(rw io.Writer, req io.Reader) command.Error {
	if c.batch != nil {
		err := c.batch.Flush()
		if err != nil {
			logutil.LogError(logger, CommandName, FlushCommandMethod, err.Error())

			return command.NewExecuteError(GetErrorCode, err)
		}
	}

	command.WriteNillableResponse(rw, &GetResponse{}, logger)

	logutil.LogDebug(logger, CommandName, FlushCommandMethod, successString)

	return nil
}
