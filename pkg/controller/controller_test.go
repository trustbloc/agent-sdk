/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package controller_test

import (
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"

	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/agent-sdk/pkg/controller"
)

func TestGetCommandHandlers(t *testing.T) {
	t.Run("test failure", func(t *testing.T) {
		ctrl, err := controller.GetCommandHandlers(&context.Provider{}, controller.WithBlocDomain("domain"))
		require.Error(t, err)
		require.Contains(t, err.Error(), api.ErrSvcNotFound.Error())
		require.Nil(t, ctrl)
	})

	t.Run("Default", func(t *testing.T) {
		framework, err := aries.New(defaults.WithInboundHTTPAddr(":26508", "", "", ""))
		require.NoError(t, err)
		require.NotNil(t, framework)

		defer func() { require.NoError(t, framework.Close()) }()

		ctx, err := framework.Context()
		require.NoError(t, err)
		require.NotNil(t, ctx)

		handlers, err := controller.GetCommandHandlers(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, handlers)
	})

	t.Run("With options", func(t *testing.T) {
		framework, err := aries.New(defaults.WithInboundHTTPAddr(":26508", "", "", ""))
		require.NoError(t, err)
		require.NotNil(t, framework)

		defer func() { require.NoError(t, framework.Close()) }()

		ctx, err := framework.Context()
		require.NoError(t, err)
		require.NotNil(t, ctx)

		handlers, err := controller.GetCommandHandlers(ctx, controller.WithBlocDomain("domain"),
			controller.WithSDSServerURL("sampleSDS"))
		require.NoError(t, err)
		require.NotEmpty(t, handlers)
	})
}
