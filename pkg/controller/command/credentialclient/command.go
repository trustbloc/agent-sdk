/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package credentialclient provides credential commands.
package credentialclient

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/internal/logutil"
)

var logger = log.New("agent-sdk-credentialclient")

const (
	// command name.
	commandName = "credentialclient"
	// command methods.
	saveCredentialCommandMethod = "SaveCredential"

	failDecodeCredentialDocDataErrMsg = "failure while decoding credential data while attempting to store credential in SDS" // nolint:lll // readability
	failStoreCredentialErrMsg         = "failure while storing credential in SDS"

	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + command.DIDClient)
)

// New returns new credential controller command instance.
func New(sdsComm *sdscomm.SDSComm) *Command {
	return &Command{
		sdsComm: sdsComm,
	}
}

// Command is controller command for credentials.
type Command struct {
	sdsComm *sdscomm.SDSComm
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []command.Handler {
	return []command.Handler{
		cmdutil.NewCommandHandler(commandName, saveCredentialCommandMethod, c.SaveCredential),
	}
}

// SaveCredential received in the request.
func (c *Command) SaveCredential(_ io.Writer, req io.Reader) command.Error {
	credentialDataToStore := sdscomm.SaveCredentialToSDSRequest{}

	err := json.NewDecoder(req).Decode(&credentialDataToStore)
	if err != nil {
		logutil.LogError(logger, commandName, saveCredentialCommandMethod,
			fmt.Sprintf("%s: %s", failDecodeCredentialDocDataErrMsg, err.Error()))

		return command.NewValidationError(InvalidRequestErrorCode,
			fmt.Errorf("%s: %w", failDecodeCredentialDocDataErrMsg, err))
	}

	return c.saveCredential(&credentialDataToStore)
}

func (c *Command) saveCredential(credentialDataToStore *sdscomm.SaveCredentialToSDSRequest) command.Error {
	err := c.sdsComm.StoreCredential(credentialDataToStore)
	if err != nil {
		logutil.LogError(logger, commandName, saveCredentialCommandMethod,
			fmt.Sprintf("%s: %s", failStoreCredentialErrMsg, err.Error()))

		return command.NewValidationError(InvalidRequestErrorCode,
			fmt.Errorf("%s: %w", failStoreCredentialErrMsg, err))
	}

	return nil
}
