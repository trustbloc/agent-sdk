/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msghandler

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/controller/webnotifier"
)

// NotificationPayload represent notification payload.
type NotificationPayload struct {
	Topic string
	Raw   []byte
}

// Notifier is implementation of channel based basic notifier.
type Notifier struct {
	connection chan<- NotificationPayload
}

// NewNotifier return notifier instance.
func NewNotifier(connection chan<- NotificationPayload) *Notifier {
	return &Notifier{connection: connection}
}

// Notify sends the given message to the subscribers.
func (n *Notifier) Notify(topic string, message []byte) error {
	msg, err := webnotifier.PrepareTopicMessage(topic, message)
	if err != nil {
		return fmt.Errorf("prepare topic message: %w", err)
	}

	n.connection <- NotificationPayload{
		Topic: topic,
		Raw:   msg,
	}

	return nil
}
