/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package zcapld provides zcapld service.
package zcapld

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/igor-pavlenko/httpsignatures-go"
	"github.com/trustbloc/edge-core/pkg/log"
	"github.com/trustbloc/edge-core/pkg/zcapld"
)

const (
	signEndpoint = "/keys/%s/sign"
)

var logger = log.New("auth-service")

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Service to provide zcapld functionality.
type Service struct {
	authzKeyStoreURL string
	userSub          string
	secretShare      string
	httpClient       httpClient
}

type signReq struct {
	Message string `json:"message,omitempty"`
}

type signResp struct {
	Signature string `json:"signature,omitempty"`
}

// New return zcap service.
func New(authzKeyStoreURL, userSub, secretShare string) *Service {
	return &Service{
		authzKeyStoreURL: authzKeyStoreURL,
		userSub:          userSub,
		secretShare:      secretShare,
		httpClient:       &http.Client{},
	}
}

// SignHeader sign header.
func (s *Service) SignHeader(req *http.Request, capabilityBytes []byte, invocationAction string) (*http.Header, error) {
	capability, err := zcapld.ParseCapability(capabilityBytes)
	if err != nil {
		return nil, err
	}

	compressedZcap, err := compressZCAP(capability)
	if err != nil {
		return nil, err
	}

	req.Header.Set(zcapld.CapabilityInvocationHTTPHeader,
		fmt.Sprintf(`zcap capability="%s",action="%s"`, compressedZcap, invocationAction))

	hs := httpsignatures.NewHTTPSignatures(&zcapld.AriesDIDKeySecrets{})
	hs.SetSignatureHashAlgorithm(&didKeySignatureHashAlgorithm{
		s: s,
	})

	err = hs.Sign(capability.Invoker, req)
	if err != nil {
		return nil, err
	}

	return &req.Header, nil
}

func compressZCAP(zcap *zcapld.Capability) (string, error) {
	raw, err := json.Marshal(zcap)
	if err != nil {
		return "", err
	}

	compressed := bytes.NewBuffer(nil)

	w := gzip.NewWriter(compressed)

	_, err = w.Write(raw)
	if err != nil {
		return "", err
	}

	err = w.Close()
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(compressed.Bytes()), nil
}

type didKeySignatureHashAlgorithm struct {
	s *Service
}

// Algorithm returns this algorithm's name.
func (a *didKeySignatureHashAlgorithm) Algorithm() string {
	return "https://github.com/hyperledger/aries-framework-go/zcaps"
}

// Create signs data with the secret.
func (a *didKeySignatureHashAlgorithm) Create(secret httpsignatures.Secret, data []byte) ([]byte, error) {
	key, err := (&zcapld.DIDKeyResolver{}).Resolve(secret.KeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve did:key URL %s: %w", secret.KeyID, err)
	}

	// TODO we are assuming curve: https://github.com/trustbloc/edge-core/issues/108.
	// TODO we shouldn't be using a `localkms` function: https://github.com/trustbloc/edge-core/issues/109.
	kid, err := localkms.CreateKID(key.Value, kms.ED25519)
	if err != nil {
		return nil, fmt.Errorf("failed to create KID from did:key: %w", err)
	}

	sig, err := a.sign(kid, data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return sig, nil
}

// Verify verifies the signature over data with the secret.
func (a *didKeySignatureHashAlgorithm) Verify(secret httpsignatures.Secret, data, signature []byte) error {
	return fmt.Errorf("not supported")
}

func (a *didKeySignatureHashAlgorithm) sign(keyID string, data []byte) ([]byte, error) {
	reqBytes, err := json.Marshal(signReq{
		Message: base64.URLEncoding.EncodeToString(data),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal create sign req : %w", err)
	}

	req, err := http.NewRequestWithContext(context.TODO(),
		http.MethodPost, a.s.authzKeyStoreURL+fmt.Sprintf(signEndpoint, keyID), bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Hub-Kms-Secret", a.s.secretShare)
	req.Header.Add("Hub-Kms-User", a.s.userSub)

	resp, _, err := sendHTTPRequest(req, a.s.httpClient, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("failed to sign from kms: %w", err)
	}

	var parsedResp signResp

	if errUnmarshal := json.Unmarshal(resp, &parsedResp); errUnmarshal != nil {
		return nil, fmt.Errorf("failed to unmarshal sign resp: %w", errUnmarshal)
	}

	signatureBytes, err := base64.URLEncoding.DecodeString(parsedResp.Signature)
	if err != nil {
		return nil, err
	}

	return signatureBytes, nil
}

func sendHTTPRequest(req *http.Request, httpClient httpClient, status int) ([]byte, http.Header, error) {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http request : %w", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			logger.Errorf("failed to close response body")
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("http request: failed to read resp body %d : %w", resp.StatusCode, err)
	}

	if resp.StatusCode != status {
		return nil, nil, fmt.Errorf("http request: expected=%d actual=%d body=%s", status, resp.StatusCode, string(body))
	}

	return body, resp.Header, nil
}
