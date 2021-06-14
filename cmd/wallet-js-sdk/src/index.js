/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

export {UniversalWallet, createWalletProfile, updateWalletProfile, contentTypes} from "./universal/vc-wallet"
export {DIDExchange, getMediatorConnections, connectToMediator, createInvitationFromRouter} from "./didcomm/connect"
export {WalletUser} from "./user/wallet-user"
export {DIDManager} from "./did/didmanager"
export {CredentialManager} from "./credential/credential-manager"
export {waitForEvent} from "./util/event"
