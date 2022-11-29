/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ariesagent //nolint:testpackage // uses internal implementation details

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/command"
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/config"
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/rest"
)

func TestNew(t *testing.T) {
	t.Run("test it creates a local agent", func(t *testing.T) {
		localAgent, err := New(&config.Options{UseLocalAgent: true})
		require.NoError(t, err)
		require.NotNil(t, localAgent)

		var a *command.Aries
		require.IsType(t, a, localAgent)
	})

	t.Run("test it creates a remote agent", func(t *testing.T) {
		remoteAgent, err := New(&config.Options{UseLocalAgent: false, AgentURL: "http://example.com"})
		require.NoError(t, err)
		require.NotNil(t, remoteAgent)

		var a *rest.Aries
		require.IsType(t, a, remoteAgent)
	})
}
