/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package command_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
)

func Test_WriteNillableResponse(t *testing.T) {
	command.WriteNillableResponse(&mockWriter{}, nil, log.New("util-test"))
}

type mockWriter struct {
}

func (m *mockWriter) Write(_ []byte) (n int, err error) {
	return 0, nil
}

type mockHandler struct {
	handle command.Exec
}

func (m *mockHandler) Name() string {
	return ""
}

func (m *mockHandler) Method() string {
	return ""
}

func (m *mockHandler) Handle() command.Exec {
	if m.handle != nil {
		return m.handle
	}

	return func(rw io.Writer, req io.Reader) command.Error {
		return nil
	}
}

func TestAriesHandler_Handle(t *testing.T) {
	t.Run("custom error", func(t *testing.T) {
		h := command.AriesHandler{Handler: &mockHandler{func(rw io.Writer, req io.Reader) command.Error {
			return command.NewExecuteError(1, errors.New("test"))
		}}}

		err := h.Handle()(nil, nil)
		require.Error(t, err)
		require.Equal(t, int(err.Code()), 1)
		require.EqualError(t, err, "test")
	})

	t.Run("success", func(t *testing.T) {
		h := command.AriesHandler{Handler: &mockHandler{}}

		err := h.Handle()(nil, nil)
		require.NoError(t, err)
	})
}
