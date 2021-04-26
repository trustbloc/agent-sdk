/*
 *
 * Copyright SecureKey Technologies Inc. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 * /
 *
 */

// Package msghandler provides messaging service related implementations.
package msghandler

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
)

// NewMessageService returns new message service instance.
func NewMessageService(name, msgType string, purpose []string, notifier command.Notifier) *MessageService {
	svc := &MessageService{
		name:     name,
		msgType:  msgType,
		purpose:  purpose,
		notifier: notifier,
	}

	return svc
}

// MessageService is basic message service implementation
// which delegates handling to registered webhook notifier.
type MessageService struct {
	name     string
	purpose  []string
	msgType  string
	notifier command.Notifier
}

// Name of the message service.
func (m *MessageService) Name() string {
	return m.name
}

// Accept matches given message type and purpose with message service type and purpose.
func (m *MessageService) Accept(msgType string, purpose []string) bool {
	purposeMatched, typeMatched := len(m.purpose) == 0, m.msgType == ""

	if purposeMatched && typeMatched {
		return false
	}

	for _, purposeCriteria := range m.purpose {
		for _, msgPurpose := range purpose {
			if purposeCriteria == msgPurpose {
				purposeMatched = true

				break
			}
		}
	}

	if m.msgType == msgType {
		typeMatched = true
	}

	return purposeMatched && typeMatched
}

// HandleInbound is inbound handler for this message service.
func (m *MessageService) HandleInbound(msg service.DIDCommMsg, ctx service.DIDCommContext) (string, error) {
	var myDID, theirDID string

	if ctx != nil {
		myDID = ctx.MyDID()
		theirDID = ctx.TheirDID()
	}

	topic := struct {
		Message  interface{} `json:"message"`
		MyDID    string      `json:"mydid"`
		TheirDID string      `json:"theirdid"`
	}{
		msg,
		myDID,
		theirDID,
	}

	bytes, err := json.Marshal(topic)
	if err != nil {
		return "", err
	}

	return "", m.notifier.Notify(m.name, bytes)
}
