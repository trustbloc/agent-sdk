/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialclient_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
	. "github.com/trustbloc/agent-sdk/pkg/controller/rest/credentialclient"
)

func TestOperation_GetRESTHandlers(t *testing.T) {
	operation := New(sdscomm.New(""))
	require.Len(t, operation.GetRESTHandlers(), 1)
}
