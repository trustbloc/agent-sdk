/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package zcapld provides zcapld service.
package zcapld //nolint:testpackage // need to mock http client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/signature"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	mockkms "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	"github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/noop"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/fingerprint"
	"github.com/igor-pavlenko/httpsignatures-go"
	"github.com/stretchr/testify/require"
)

func TestService_SignHeader(t *testing.T) {
	t.Run("test wrong capability", func(t *testing.T) {
		svc := New("")

		hdr, err := svc.SignHeader(&http.Request{Header: make(map[string][]string)}, []byte(""))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal zcap")
		require.Nil(t, hdr)
	})

	t.Run("test error creating signature", func(t *testing.T) {
		svc := New("")

		hdr, err := svc.SignHeader(&http.Request{
			Method: http.MethodGet,
			Header: make(map[string][]string),
		}, []byte("{}"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "error creating signature")
		require.Nil(t, hdr)
	})
}

func TestDidKeySignatureHashAlgorithm_Create(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		svc := New("")
		svc.httpClient = &mockHTTPClient{DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf(`{"signature":"%s"}`,
					base64.URLEncoding.EncodeToString([]byte("value")))))),
			}, nil
		}}

		s := testSigner(t)

		_, didKeyURL := fingerprint.CreateDIDKey(s.PublicKeyBytes())

		didKeySig := didKeySignatureHashAlgorithm{s: svc}

		_, err := didKeySig.Create(httpsignatures.Secret{KeyID: didKeyURL}, []byte("data"))
		require.NoError(t, err)
	})

	t.Run("test wrong signature", func(t *testing.T) {
		svc := New("")
		svc.httpClient = &mockHTTPClient{DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"signature":"1"}`))),
			}, nil
		}}

		s := testSigner(t)

		_, didKeyURL := fingerprint.CreateDIDKey(s.PublicKeyBytes())

		didKeySig := didKeySignatureHashAlgorithm{s: svc}

		_, err := didKeySig.Create(httpsignatures.Secret{KeyID: didKeyURL}, []byte("data"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "illegal base64 data at input byte 0")
	})

	t.Run("test failed to unmarshal resp", func(t *testing.T) {
		svc := New("")
		svc.httpClient = &mockHTTPClient{DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
			}, nil
		}}

		s := testSigner(t)

		_, didKeyURL := fingerprint.CreateDIDKey(s.PublicKeyBytes())

		didKeySig := didKeySignatureHashAlgorithm{s: svc}

		_, err := didKeySig.Create(httpsignatures.Secret{KeyID: didKeyURL}, []byte("data"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal sign resp")
	})

	t.Run("test resp is not 200", func(t *testing.T) {
		svc := New("")
		svc.httpClient = &mockHTTPClient{DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
			}, nil
		}}

		s := testSigner(t)

		_, didKeyURL := fingerprint.CreateDIDKey(s.PublicKeyBytes())

		didKeySig := didKeySignatureHashAlgorithm{s: svc}

		_, err := didKeySig.Create(httpsignatures.Secret{KeyID: didKeyURL}, []byte("data"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected=200 actual=500")
	})

	t.Run("test error from http client", func(t *testing.T) {
		svc := New("")
		svc.httpClient = &mockHTTPClient{DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("http client error")
		}}

		s := testSigner(t)

		_, didKeyURL := fingerprint.CreateDIDKey(s.PublicKeyBytes())

		didKeySig := didKeySignatureHashAlgorithm{s: svc}

		_, err := didKeySig.Create(httpsignatures.Secret{KeyID: didKeyURL}, []byte("data"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "http client error")
	})
}

func TestDidKeySignatureHashAlgorithm_Verify(t *testing.T) {
	didKeySig := didKeySignatureHashAlgorithm{}
	err := didKeySig.Verify(httpsignatures.Secret{}, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not supported")
}

func testSigner(t *testing.T) signature.Signer {
	t.Helper()

	k, err := localkms.New(
		"local-lock://custom/master/key/",
		mockkms.NewProviderForKMS(storage.NewMockStoreProvider(), &noop.NoLock{}),
	)
	require.NoError(t, err)

	tc, err := tinkcrypto.New()
	require.NoError(t, err)

	s, err := signature.NewCryptoSigner(tc, k, "ED25519")
	require.NoError(t, err)

	return s
}

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}
