/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package store provides store commands.
package store

import (
	"encoding/json"
	"io"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/hyperledger/aries-framework-go/spi/storage"
	"github.com/trustbloc/edge-core/pkg/log"

	agentcmd "github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/logutil"
)

const (
	// CommandName package command name.
	CommandName = "store"
	// PutCommandMethod command method.
	PutCommandMethod = "Put"
	// GetCommandMethod command method.
	GetCommandMethod = "Get"
	// QueryCommandMethod command method.
	QueryCommandMethod = "Query"
	// DeleteCommandMethod command method.
	DeleteCommandMethod = "Delete"
	// FlushCommandMethod command method.
	FlushCommandMethod = "Flush"

	successString = "success"
)

const (
	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + agentcmd.Store)
	// PutErrorCode is typically a code for Put errors.
	PutErrorCode
	// GetErrorCode is typically a code for Get errors.
	GetErrorCode
	// QueryErrorCode is typically a code for Query errors.
	QueryErrorCode
	// DeleteErrorCode is typically a code for Delete errors.
	DeleteErrorCode
	// FlushErrorCode is typically a code for Flush errors.
	FlushErrorCode
)

var logger = log.New("agent-sdk-store")

// Provider describes dependencies for the client.
type Provider interface {
	StorageProvider() storage.Provider
}

// Command is controller command for store.
type Command struct {
	provider storage.Provider
	store    storage.Store
}

// New returns new store controller command instance.
func New(p Provider) (*Command, error) {
	store, err := p.StorageProvider().OpenStore(CommandName)
	if err != nil {
		return nil, err
	}

	return &Command{provider: p.StorageProvider(), store: store}, nil
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []command.Handler {
	return []command.Handler{
		cmdutil.NewCommandHandler(CommandName, PutCommandMethod, c.Put),
		cmdutil.NewCommandHandler(CommandName, GetCommandMethod, c.Get),
		cmdutil.NewCommandHandler(CommandName, QueryCommandMethod, c.Query),
		cmdutil.NewCommandHandler(CommandName, DeleteCommandMethod, c.Delete),
		cmdutil.NewCommandHandler(CommandName, FlushCommandMethod, c.Flush),
	}
}

// Put stores the key, value and (optional) tags.
func (c *Command) Put(rw io.Writer, req io.Reader) command.Error {
	var request PutRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, PutCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	if err = c.store.Put(request.Key, request.Value, request.Tags...); err != nil {
		logutil.LogError(logger, CommandName, PutCommandMethod, err.Error())

		return command.NewExecuteError(PutErrorCode, err)
	}

	command.WriteNillableResponse(rw, nil, logger)

	logutil.LogDebug(logger, CommandName, PutCommandMethod, successString)

	return nil
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

// Query retrieves data according to given expression.
func (c *Command) Query(rw io.Writer, req io.Reader) command.Error {
	var request QueryRequest

	err := json.NewDecoder(req).Decode(&request)
	if err != nil {
		logutil.LogError(logger, CommandName, QueryCommandMethod, err.Error())

		return command.NewValidationError(InvalidRequestErrorCode, err)
	}

	var options []storage.QueryOption

	if request.PageSize > 0 {
		options = append(options, storage.WithPageSize(request.PageSize))
	}

	records, err := c.store.Query(request.Expression, options...)
	if err != nil {
		logutil.LogError(logger, CommandName, QueryCommandMethod, err.Error())

		return command.NewExecuteError(QueryErrorCode, err)
	}

	defer func() {
		errClose := records.Close()
		if errClose != nil {
			logutil.LogError(logger, CommandName, QueryCommandMethod, errClose.Error())
		}
	}()

	values, cmdErr := getValuesFromIterator(records)
	if cmdErr != nil {
		return cmdErr
	}

	command.WriteNillableResponse(rw, &QueryResponse{Results: values}, logger)

	logutil.LogDebug(logger, CommandName, QueryCommandMethod, successString)

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

// Flush data in all currently open stores.
func (c *Command) Flush(rw io.Writer, _ io.Reader) command.Error {
	openStores := c.provider.GetOpenStores()

	for _, openStore := range openStores {
		err := openStore.Flush()
		if err != nil {
			logutil.LogError(logger, CommandName, FlushCommandMethod, err.Error())

			return command.NewExecuteError(FlushErrorCode, err)
		}
	}

	command.WriteNillableResponse(rw, &GetResponse{}, logger)

	logutil.LogDebug(logger, CommandName, FlushCommandMethod, successString)

	return nil
}

func getValuesFromIterator(iterator storage.Iterator) ([][]byte, command.Error) {
	var values [][]byte

	moreRecords, err := iterator.Next()
	if err != nil {
		logutil.LogError(logger, CommandName, QueryCommandMethod, err.Error())

		return nil, command.NewExecuteError(QueryErrorCode, err)
	}

	for moreRecords {
		value, errValue := iterator.Value()
		if errValue != nil {
			logutil.LogError(logger, CommandName, QueryCommandMethod, errValue.Error())

			return nil, command.NewExecuteError(QueryErrorCode, errValue)
		}

		values = append(values, value)

		var errNext error

		moreRecords, errNext = iterator.Next()
		if errNext != nil {
			logutil.LogError(logger, CommandName, QueryCommandMethod, errNext.Error())

			return nil, command.NewExecuteError(QueryErrorCode, errNext)
		}
	}

	return values, nil
}
