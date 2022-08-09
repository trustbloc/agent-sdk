/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package command

import "github.com/hyperledger/aries-framework-go/pkg/controller/command"

const (
	// ValidationError is error type for command validation errors.
	ValidationError command.Type = iota
	// ExecuteError is error type for command execution failure.
	ExecuteError command.Type = iota
)

const (
	// UnknownStatus default error code for unknown errors.
	UnknownStatus command.Code = iota
)

// Group is the error groups.
// Note: recommended to use [0-9]*000 pattern for any new entries.
// Example: 2000, 3000, 4000 ...... 25000.
type Group int32

const (
	// DIDClient error group for DID client command errors.
	DIDClient Group = 1000

	// MediatorClient error group for mediator client command errors.
	MediatorClient Group = 2000

	// Store error group for Store command errors.
	Store Group = 3000
)

// NewExecuteError returns new command execute error.
func NewExecuteError(code command.Code, err error) command.Error {
	return &commandError{err, code, ExecuteError}
}

// NewValidationError returns new command validation error.
func NewValidationError(code command.Code, err error) command.Error {
	return &commandError{err, code, ValidationError}
}

// commandError implements basic command Error.
type commandError struct {
	error
	code    command.Code
	errType command.Type
}

func (c *commandError) Code() command.Code {
	return c.code
}

func (c *commandError) Type() command.Type {
	return c.errType
}
