/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package agent-rest (Agent SDK REST Server) of trustbloc/agent-sdk.
//
// Terms Of Service:
//
//	Schemes: https
//	Version: 0.1.0
//	License: SPDX-License-Identifier: Apache-2.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package main

import (
	"github.com/hyperledger/aries-framework-go/pkg/common/log"
	"github.com/spf13/cobra"

	"github.com/trustbloc/agent-sdk/cmd/agent-rest/startcmd"
)

// This is an application which starts Aries agent controller API on given port.
func main() {
	rootCmd := &cobra.Command{
		Use: "agent-rest",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	logger := log.New("agent-sdk/agent-rest")

	startCmd, err := startcmd.Cmd(&startcmd.HTTPServer{})
	if err != nil {
		logger.Fatalf(err.Error())
	}

	rootCmd.AddCommand(startCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("Failed to run aries-agent-rest: %s", err)
	}
}
