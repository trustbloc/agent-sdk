[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/trustbloc/edge-agent/main/LICENSE)

## Wallet SDK Data Models

Wallet SDK supports all types of [Aries Verifiable Credential Wallet data models](https://github.com/hyperledger/aries-framework-go/blob/main/docs/vc_wallet.md#supported-data-models).

Wallet SDK has introduced few customized data models to support features like saving user preferences, blinded routing, collections.

| Data Model Type|  Model   |   JSON-LD Context |   Purpose |
|   --- | ---  | --- | --- |
|   Metadata    |   User Preference |   https://trustbloc.github.io/context/wallet/user-preferences-v1.jsonld   |    This data model is mainly used by wallet user module to save user specific wallet settings like display features like name, image, description and for settings related digital proof generation like controller, signature type, verification method etc. Refer [wallet user API reference](wallet_sdk.md#module_wallet-user) for more details|
|   Metadata    |   Manifest Credential Mapping |   https://trustbloc.github.io/context/wallet/manifest-mapping-v1.jsonld   |   User by blinded routing feature to map a given manifest credential to a connection. Refer [blinded routing API reference](wallet_sdk.md#module_blinded-routing) for more details.    |
|   Collection    |  Content Vault   |   https://trustbloc.github.io/context/wallet/collections-v1.jsonld    |  Used for grouping wallet contents by collection. For example: credential vaults. Refer [collection API reference](wallet_sdk.md#module_collection) for more details. |


## Contributing
Thank you for your interest in contributing. Please see our [community contribution guidelines](https://github.com/trustbloc/community/blob/main/CONTRIBUTING.md) for more information.

## License
Apache License, Version 2.0 (Apache-2.0). See the [LICENSE](LICENSE) file.
