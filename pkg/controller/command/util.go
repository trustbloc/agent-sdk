/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package command

import (
	"encoding/json"
	"io"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"
)

type logger interface {
	Errorf(string, ...interface{})
}

// WriteNillableResponse is a utility function that writes v to w.
// If v is nil then an empty object is written.
// TODO this capability should be injected into the command implementations.
func WriteNillableResponse(w io.Writer, v interface{}, l logger) {
	obj := v
	if v == nil {
		obj = map[string]interface{}{}
	}
	// TODO as of now, just log errors for writing response
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		l.Errorf("Unable to send error response, %s", err)
	}
}

// AriesHandler implements aries handler.
type AriesHandler struct {
	Handler
}

// Handle execute function of the command.
func (ah AriesHandler) Handle() command.Exec {
	return func(rw io.Writer, req io.Reader) command.Error {
		if err := ah.Handler.Handle()(rw, req); err != nil {
			return command.NewExecuteError(command.Code(err.Code()), err)
		}

		return nil
	}
}
