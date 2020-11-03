/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package didclient // nolint:testpackage // uses internal implementation details

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	mediatorsvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	mockprotocol "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol"
	mockroute "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/mediator"
	mockvdr "github.com/hyperledger/aries-framework-go/pkg/mock/vdr"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/edv/pkg/edvprovider/memedvprovider"
	"github.com/trustbloc/edv/pkg/restapi"
	didclient "github.com/trustbloc/trustbloc-did-method/pkg/did"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
)

func TestNew(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)
		require.NotNil(t, c.GetHandlers())
	})

	t.Run("test no coordination service error", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, &mockprotocol.MockProvider{
			ServiceErr: fmt.Errorf("sample-error"),
		})
		require.Error(t, err)
		require.Nil(t, c)
		require.EqualError(t, err, "sample-error")
	})

	t.Run("test invalid coordination service error", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, &mockprotocol.MockProvider{
			ServiceMap: map[string]interface{}{
				mediatorsvc.Coordination: "xyz",
			},
		})
		require.Error(t, err)
		require.Nil(t, c)
		require.EqualError(t, err, "cast service to route service failed")
	})
}

func TestCommand_CreateBlocDID(t *testing.T) {
	t.Run("test error from request", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBufferString("--"))
		require.Error(t, cmdErr)
		require.Equal(t, InvalidRequestErrorCode, cmdErr.Code())
		require.Equal(t, command.ValidationError, cmdErr.Type())
	})

	t.Run("test error from create did", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		req, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{{
			ID: "key1", Type: "key1",
			Value: base64.RawURLEncoding.EncodeToString([]byte("value")),
		}}})
		require.NoError(t, err)

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "error create did")
	})

	t.Run("test error from did base64 decode", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.didBlocClient = &mockDIDClient{createDIDErr: fmt.Errorf("error create did")}

		var b bytes.Buffer

		req, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{{
			ID: "key1", Type: "key1",
			Value: "value",
		}}})
		require.NoError(t, err)

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(req))
		require.Error(t, cmdErr)
		require.Equal(t, CreateDIDErrorCode, cmdErr.Code())
		require.Equal(t, command.ExecuteError, cmdErr.Type())
		require.Contains(t, cmdErr.Error(), "illegal base64 data")
	})

	c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
	require.NoError(t, err)
	require.NotNil(t, c)

	c.didBlocClient = &mockDIDClient{createDIDValue: &did.Doc{ID: "1"}}

	var b bytes.Buffer

	t.Run("test success create did with Ed25519 key", func(t *testing.T) {
		// ED key
		r, err := json.Marshal(CreateBlocDIDRequest{PublicKeys: []PublicKey{{
			ID: "key1", Type: "key1", KeyType: "Ed25519",
			Value: base64.RawURLEncoding.EncodeToString([]byte("value")),
		}}})
		require.NoError(t, err)

		cmdErr := c.CreateTrustBlocDID(&b, bytes.NewBuffer(r))
		require.NoError(t, cmdErr)

		resp := &CreateDIDResponse{}
		err = json.NewDecoder(&b).Decode(&resp)
		require.NoError(t, err)

		var didMap map[string]string
		err = json.Unmarshal(resp.DID, &didMap)
		require.NoError(t, err)

		require.Equal(t, "1", didMap["id"])
	})
}

func TestCommand_CreatePeerDID(t *testing.T) {
	t.Run("test error from request", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString("--"))
		require.Error(t, cmdErr)
		require.Equal(t, InvalidRequestErrorCode, cmdErr.Code())
		require.Equal(t, command.ValidationError, cmdErr.Type())

		cmdErr = c.CreatePeerDID(&b, bytes.NewBufferString("{}"))
		require.Error(t, cmdErr)
		require.Equal(t, InvalidRequestErrorCode, cmdErr.Code())
		require.Equal(t, command.ValidationError, cmdErr.Type())
	})

	t.Run("success (registered route)", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		routerEndpoint := "http://router.com"
		keys := []string{"abc", "xyz"}
		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:              uuid.New().String(),
					Type:            didCommServiceType,
					ServiceEndpoint: routerEndpoint,
					RoutingKeys:     keys,
					RecipientKeys:   []string{"1ert5", "x5356s"},
				},
			},
		}}

		mediatorConfig := mediatorsvc.NewConfig(routerEndpoint, keys)
		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return mediatorConfig, nil
			},
		}

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Nil(t, cmdErr)

		resp := &CreateDIDResponse{}
		err = json.NewDecoder(&b).Decode(&resp)
		require.NoError(t, err)

		var didMap map[string]interface{}
		err = json.Unmarshal(resp.DID, &didMap)
		require.NoError(t, err)
		require.NotEmpty(t, didMap["service"])
	})

	t.Run("success (default)", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:              uuid.New().String(),
					Type:            didCommServiceType,
					ServiceEndpoint: "http://router.com",
				},
			},
		}}

		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return &mediatorsvc.Config{}, nil
			},
		}

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Nil(t, cmdErr)

		resp := &CreateDIDResponse{}
		err = json.NewDecoder(&b).Decode(&resp)
		require.NoError(t, err)

		var didMap map[string]interface{}
		err = json.Unmarshal(resp.DID, &didMap)
		require.NoError(t, err)
		require.NotEmpty(t, didMap["service"])
	})

	t.Run("test error while creating peer DID", func(t *testing.T) {
		c, err := New("domain", &sdscomm.SDSComm{}, getMockProvider())
		require.NoError(t, err)
		require.NotNil(t, c)

		// error while getting mediator config
		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return nil, fmt.Errorf("sample-error-1")
			},
		}

		var b bytes.Buffer

		cmdErr := c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Error(), "sample-error-1")

		// error while create peer DID from vdri
		c.mediatorClient = &mockMediatorClient{
			GetConfigFunc: func(connID string) (*mediatorsvc.Config, error) {
				return &mediatorsvc.Config{}, nil
			},
		}

		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateErr: fmt.Errorf("sample-error-2")}

		cmdErr = c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Error(), "sample-error-2")

		// error for missing 'did-communication'
		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			Context: []string{"https://w3id.org/did/v1"},
		}}

		cmdErr = c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Error(), fmt.Sprintf(errMissingDIDCommServiceType, didCommServiceType))

		// error while adding router key
		c.vdrRegistry = &mockvdr.MockVDRegistry{CreateValue: &did.Doc{
			Context: []string{"https://w3id.org/did/v1"},
			Service: []did.Service{
				{
					ID:              uuid.New().String(),
					Type:            didCommServiceType,
					ServiceEndpoint: "http://router.com",
					RecipientKeys:   []string{"1ert5", "x5356s"},
				},
			},
		}}

		c.mediatorSvc = &mockroute.MockMediatorSvc{
			AddKeyErr: fmt.Errorf("sample-error-3"),
		}

		cmdErr = c.CreatePeerDID(&b, bytes.NewBufferString(`{
			"routerConnectionID" : "abcd-sample-id"
		}`))

		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), "sample-error-3")
	})
}

func TestCommand_SaveDID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sdsSrv := newTestEDVServer(t)
		defer sdsSrv.Close()

		sampleDIDDocData := sdscomm.SaveDIDDocToSDSRequest{}

		didDocDataBytes, err := json.Marshal(sampleDIDDocData)
		require.NoError(t, err)

		sdsComm := sdscomm.New(fmt.Sprintf("%s/encrypted-data-vaults", sdsSrv.URL))

		cmd, err := New("", sdsComm, getMockProvider())
		require.NoError(t, err)

		cmdErr := cmd.SaveDID(nil, bytes.NewBuffer(didDocDataBytes))
		require.NoError(t, cmdErr)
	})
	t.Run("Fail to unmarshal - invalid SaveDIDDocToSDSRequest", func(t *testing.T) {
		cmd, err := New("", sdscomm.New("SomeURL"), getMockProvider())
		require.NoError(t, err)
		cmdErr := cmd.SaveDID(nil, bytes.NewBuffer([]byte("")))
		require.Contains(t, cmdErr.Error(), errDecodeDIDDocDataErrMsg)
	})
	t.Run("Fail to save DID document - bad SDS server URL", func(t *testing.T) {
		cmd, err := New("", sdscomm.New("BadURL"), getMockProvider())
		require.NoError(t, err)

		sampleDIDDocData := sdscomm.SaveDIDDocToSDSRequest{}

		didDocDataBytes, err := json.Marshal(sampleDIDDocData)
		require.NoError(t, err)

		cmdErr := cmd.SaveDID(nil, bytes.NewBuffer(didDocDataBytes))
		require.Contains(t, cmdErr.Error(), `failure while storing DID document in SDS: failure while `+
			`ensuring that the user's DID vault exists: unexpected error during the "create vault" call `+
			`to SDS: failed to send POST request:`)
	})
}

type mockDIDClient struct {
	createDIDValue *did.Doc
	createDIDErr   error
}

func (m *mockDIDClient) CreateDID(domain string, opts ...didclient.CreateDIDOption) (*did.Doc, error) {
	return m.createDIDValue, m.createDIDErr
}

// mockMediatorClient mock mediator client.
type mockMediatorClient struct {
	RegisterErr   error
	GetConfigFunc func(connID string) (*mediatorsvc.Config, error)
}

// Register registers with the router.
func (c *mockMediatorClient) Register(connectionID string) error {
	if c.RegisterErr != nil {
		return c.RegisterErr
	}

	return nil
}

// GetConfig gets the router config.
func (c *mockMediatorClient) GetConfig(connID string) (*mediatorsvc.Config, error) {
	return c.GetConfigFunc(connID)
}

func newTestEDVServer(t *testing.T) *httptest.Server {
	edvService, err := restapi.New(memedvprovider.NewProvider())
	require.NoError(t, err)

	handlers := edvService.GetOperations()
	router := mux.NewRouter()
	router.UseEncodedPath()

	for _, handler := range handlers {
		router.HandleFunc(handler.Path(), handler.Handle()).Methods(handler.Method())
	}

	return httptest.NewServer(router)
}

func getMockProvider() Provider {
	return &mockprotocol.MockProvider{
		ServiceMap: map[string]interface{}{
			mediatorsvc.Coordination: &mockroute.MockMediatorSvc{},
		},
	}
}
