/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command // nolint:testpackage // uses internal implementation details

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command/didexchange"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/config"
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

func TestNewAries(t *testing.T) {
	t.Run("test it creates an instance with a framework and handlers", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)
	})
}

type handlerFunc func(topic string, message []byte) error

func (hf handlerFunc) Handle(topic string, message []byte) error {
	return hf(topic, message)
}

func TestAries_RegisterHandler(t *testing.T) {
	const topic = "didexchange_states"

	a, err := NewAries(&config.Options{})
	require.NoError(t, err)
	require.NotNil(t, a)

	done := make(chan struct{})

	defer a.UnregisterHandler(a.RegisterHandler(handlerFunc(func(topic string, message []byte) error {
		if strings.Contains(string(message), "post_state") {
			close(done)

			return nil
		}

		return errors.New("error")
	}), topic))

	ctrl, err := a.GetDIDExchangeController()
	require.NoError(t, err)

	inv := ctrl.CreateInvitation(&models.RequestEnvelope{Payload: []byte(`{}`)})

	var resp *didexchange.CreateInvitationResponse

	require.NoError(t, json.Unmarshal(inv.Payload, &resp))

	src, err := json.Marshal(resp.Invitation)
	require.NoError(t, err)

	ctrl.ReceiveInvitation(&models.RequestEnvelope{Payload: src})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("timeout")
	}
}

func TestCreateVDRs(t *testing.T) {
	tests := []struct {
		name              string
		resolvers         []string
		blocDomain        string
		trustblocResolver string
		expected          int
		accept            map[int][]string
	}{{
		name: "Empty data",
		// expects default trustbloc resolver
		accept:   map[int][]string{0: {"orb"}},
		expected: 1,
	}, {
		name:      "Groups methods by resolver",
		resolvers: []string{"orb@http://resolver.com", "v1@http://resolver.com"},
		accept:    map[int][]string{0: {"orb", "v1"}, 1: {"orb"}},
		// expects resolver.com that supports trustbloc,v1 methods and default trustbloc resolver
		expected: 2,
	}, {
		name:      "Two different resolvers",
		resolvers: []string{"orb@http://resolver1.com", "v1@http://resolver2.com"},
		accept:    map[int][]string{0: {"orb"}, 1: {"v1"}, 2: {"orb"}},
		// expects resolver1.com and resolver2.com that supports trustbloc and v1 methods and default trustbloc resolver
		expected: 3,
	}}

	for _, test := range tests {
		res, err := createVDRs(test.resolvers, test.blocDomain)

		for i, methods := range test.accept {
			for _, method := range methods {
				require.True(t, res[i].Accept(method))
			}
		}

		require.NoError(t, err)
		require.Equal(t, test.expected, len(res))
	}
}

func TestAries_GetIntroduceController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		ic, err := a.GetIntroduceController()
		require.NoError(t, err)
		require.NotNil(t, ic)
	})
}

func TestAries_GetVerifiableController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		ic, err := a.GetVerifiableController()
		require.NoError(t, err)
		require.NotNil(t, ic)
	})
}

func TestAries_GetDIDExchangeController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		dec, err := a.GetDIDExchangeController()
		require.NoError(t, err)
		require.NotNil(t, dec)
	})
}

func TestAries_GetIssueCredentialController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		ic, err := a.GetIssueCredentialController()
		require.NoError(t, err)
		require.NotNil(t, ic)
	})
}

func TestAries_GetPresentProofController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		p, err := a.GetPresentProofController()
		require.NoError(t, err)
		require.NotNil(t, p)
	})
}

func TestAries_GetVDRController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		v, err := a.GetVDRController()
		require.NoError(t, err)
		require.NotNil(t, v)
	})
}

func TestAries_GetMediatorController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		m, err := a.GetMediatorController()
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestAries_GetMessagingController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		m, err := a.GetMessagingController()
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestAries_GetOutOfBandController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		controller, err := a.GetOutOfBandController()
		require.NoError(t, err)
		require.NotNil(t, controller)
	})
}

func TestAries_GetKMSController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		controller, err := a.GetKMSController()
		require.NoError(t, err)
		require.NotNil(t, controller)
	})
}

func TestAries_GetMediatorClientController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		m, err := a.GetMediatorClientController()
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestAries_GetBlindedRoutingController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		m, err := a.GetBlindedRoutingController()
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}

func TestAries_GetVCWalletController(t *testing.T) {
	t.Run("it creates a controller", func(t *testing.T) {
		opts := &config.Options{}
		a, err := NewAries(opts)
		require.NoError(t, err)
		require.NotNil(t, a)

		m, err := a.GetVCWalletController()
		require.NoError(t, err)
		require.NotNil(t, m)
	})
}
