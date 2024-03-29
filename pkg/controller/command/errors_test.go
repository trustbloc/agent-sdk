/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package command_test

import (
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/stretchr/testify/require"

	agencmd "github.com/trustbloc/agent-sdk/pkg/controller/command"
)

func TestNewExecuteError(t *testing.T) {
	e := agencmd.NewExecuteError(1, fmt.Errorf("error"))
	require.Equal(t, "error", e.Error())
	require.Equal(t, command.Code(1), e.Code())
	require.Equal(t, command.Type(1), e.Type())
}

func TestNewValidationError(t *testing.T) {
	e := agencmd.NewValidationError(1, fmt.Errorf("error"))
	require.Equal(t, "error", e.Error())
	require.Equal(t, command.Code(1), e.Code())
	require.Equal(t, command.Type(0), e.Type())
}
