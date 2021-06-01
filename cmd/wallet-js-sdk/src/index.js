/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

export  {UniversalWallet, createWalletProfile, updateWalletProfile} from "./universal/vc-wallet"
export  {DIDExchange, getMediatorConnections, connectToMediator, createInvitationFromRouter} from "./didcomm/connect"
export {waitForEvent} from "./util/event"
