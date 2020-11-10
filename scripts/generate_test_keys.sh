#!/bin/sh
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Generating Aries-Framework-Go Test PKI"
cd /opt/go/src/github.com/trustbloc/agent-sdk
mkdir -p deployments/keys/tls
tmp=$(mktemp)
echo "subjectKeyIdentifier=hash
authorityKeyIdentifier = keyid,issuer
extendedKeyUsage = serverAuth
keyUsage = Digital Signature, Key Encipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
DNS.2 = alice.agent.sdk.example.com
DNS.3 = bob.agent.sdk.example.com
DNS.4 = carl.agent.sdk.example.com
DNS.5 = carl.router.agent.sdk.example.com" >> "$tmp"

#create CA
openssl ecparam -name prime256v1 -genkey -noout -out deployments/keys/tls/ec-cakey.pem
openssl req -new -x509 -key deployments/keys/tls/ec-cakey.pem -subj "/C=CA/ST=ON/O=Example Internet CA Inc.:CA Sec/OU=CA Sec" -out deployments/keys/tls/ec-cacert.pem

#create TLS creds
openssl ecparam -name prime256v1 -genkey -noout -out deployments/keys/tls/ec-key.pem
openssl req -new -key deployments/keys/tls/ec-key.pem -subj "/C=CA/ST=ON/O=Example Inc.:Aries-Framework-Go/OU=Aries-Framework-Go/CN=*.example.com" -out deployments/keys/tls/ec-key.csr
openssl x509 -req -in deployments/keys/tls/ec-key.csr -CA deployments/keys/tls/ec-cacert.pem -CAkey deployments/keys/tls/ec-cakey.pem -CAcreateserial -extfile "$tmp" -out deployments/keys/tls/ec-pubCert.pem -days 365


echo "done generating Agent SDK PKI"
