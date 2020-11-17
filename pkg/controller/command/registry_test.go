/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package command_test

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
)

func TestRegistry_Execute(t *testing.T) {
	handlers := []command.Handler{
		cmdutil.NewCommandHandler("SampleCommand", "SampleMethod", func(rw io.Writer, req io.Reader) command.Error {
			var request struct {
				A int
				B int
			}

			err := json.NewDecoder(req).Decode(&request)
			if err != nil {
				return command.NewExecuteError(9999, err)
			}

			err = json.NewEncoder(rw).Encode(&struct {
				Result int
			}{
				Result: request.A + request.B,
			})
			if err != nil {
				return command.NewExecuteError(99999, err)
			}

			return nil
		},
		),

		cmdutil.NewCommandHandler("SampleCommand", "ProblematicMethod", func(rw io.Writer, req io.Reader) command.Error {
			return command.NewExecuteError(99999, fmt.Errorf("sample-error"))
		}),
	}

	t.Run("Successful execution of command", func(t *testing.T) {
		registry := command.NewRegistry(handlers)

		res := struct {
			Result int
		}{}

		err := registry.Execute("SampleCommand", "SampleMethod", struct {
			A int
			B int
		}{
			A: 10,
			B: 21,
		}, &res)

		require.NoError(t, err)
		require.Equal(t, res.Result, 31)
	})

	t.Run("Command execution failure", func(t *testing.T) {
		registry := command.NewRegistry(handlers)
		err := registry.Execute("SampleCommand", "ProblematicMethod", struct {
			A int
			B int
		}{
			A: 10,
			B: 21,
		}, &struct{}{})

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to execute command: sample-error")
	})

	t.Run("Command not found", func(t *testing.T) {
		registry := command.NewRegistry(handlers)
		err := registry.Execute("InvalidCommand", "SampleMethod", struct {
			A int
			B int
		}{
			A: 10,
			B: 21,
		}, &struct{}{})

		require.Error(t, err)
		require.Contains(t, err.Error(), "could not find matching registered handler")
	})

	t.Run("invalid request type", func(t *testing.T) {
		registry := command.NewRegistry(handlers)
		err := registry.Execute("SampleCommand", "SampleMethod", make(chan int), &struct{}{})

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read command request")
	})

	t.Run("invalid response type", func(t *testing.T) {
		registry := command.NewRegistry(handlers)
		err := registry.Execute("SampleCommand", "SampleMethod", struct {
			A int
			B int
		}{
			A: 10,
			B: 21,
		}, make(chan int))

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get command response")
	})
}
