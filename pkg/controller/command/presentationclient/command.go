/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package presentationclient

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/agent-sdk/pkg/controller/command"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/internal/cmdutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/internal/logutil"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
)

var logger = log.New("agent-sdk-presentationclient")

const (
	// command name.
	commandName = "presentationclient"
	// command methods.
	savePresentationCommandMethod = "SavePresentation"

	failDecodePresentationDocDataErrMsg = "failure while decoding presentation data while attempting to store presentation in SDS" // nolint:lll // readability
	failStorePresentationErrMsg         = "failure while storing presentation in SDS"

	// InvalidRequestErrorCode is typically a code for validation errors.
	InvalidRequestErrorCode = command.Code(iota + command.DIDClient)
)

// New returns new presentation controller command instance.
func New(sdsComm *sdscomm.SDSComm) *Command {
	return &Command{
		sdsComm: sdsComm,
	}
}

// Command is controller command for presentations.
type Command struct {
	sdsComm *sdscomm.SDSComm
}

// GetHandlers returns list of all commands supported by this controller command.
func (c *Command) GetHandlers() []command.Handler {
	return []command.Handler{
		cmdutil.NewCommandHandler(commandName, savePresentationCommandMethod, c.SavePresentation),
	}
}

// SavePresentation in the SDS.
func (c *Command) SavePresentation(_ io.Writer, req io.Reader) command.Error {
	presentationDataToStore := sdscomm.SavePresentationToSDSRequest{}

	err := json.NewDecoder(req).Decode(&presentationDataToStore)
	if err != nil {
		logutil.LogError(logger, commandName, savePresentationCommandMethod,
			fmt.Sprintf("%s: %s", failDecodePresentationDocDataErrMsg, err.Error()))

		return command.NewValidationError(InvalidRequestErrorCode,
			fmt.Errorf("%s: %w", failDecodePresentationDocDataErrMsg, err))
	}

	return c.savePresentation(&presentationDataToStore)
}

func (c *Command) savePresentation(presentationDataToStore *sdscomm.SavePresentationToSDSRequest) command.Error {
	err := c.sdsComm.StorePresentation(presentationDataToStore)
	if err != nil {
		logutil.LogError(logger, commandName, savePresentationCommandMethod,
			fmt.Sprintf("%s: %s", failStorePresentationErrMsg, err.Error()))

		return command.NewValidationError(InvalidRequestErrorCode,
			fmt.Errorf("%s: %w", failStorePresentationErrMsg, err))
	}

	return nil
}
