/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Registry is collection of all the commands which can be used for communication between different
// command libraries.
type Registry struct {
	handlers []Handler
}

// NewRegistry return new command registry instance.
func NewRegistry(handlers []Handler) *Registry {
	return &Registry{handlers: handlers}
}

// Execute finds command by name and method, and then executes it.
func (n *Registry) Execute(name, method string, req, res interface{}) error {
	handler := n.findHandler(name, method)
	if handler == nil {
		return fmt.Errorf("could not find matching registered handler for given command '%s'", name)
	}

	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to read command request: %w", err)
	}

	var rw bytes.Buffer

	err = handler.Handle()(&rw, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	if res != nil {
		err = json.Unmarshal(rw.Bytes(), res)
		if err != nil {
			return fmt.Errorf("failed to get command response: %w", err)
		}
	}

	return nil
}

func (n *Registry) findHandler(name, method string) Handler {
	if len(n.handlers) == 0 {
		return nil
	}

	for _, h := range n.handlers {
		if h.Name() == name && h.Method() == method {
			return h
		}
	}

	return nil
}
