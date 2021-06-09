/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

export {UniversalWallet, createWalletProfile, updateWalletProfile, contentTypes} from "./universal/vc-wallet"
export {DIDExchange, getMediatorConnections, connectToMediator, createInvitationFromRouter} from "./didcomm/connect"
export {WalletUser} from "./user/walletUser"
export {DIDManager} from "./did/didmanager"
export {waitForEvent} from "./util/event"
