[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/trustbloc/edge-agent/main/LICENSE)

# Wallet SDK

JavaScript SDK based on [Aries Verifiable Credential](https://github.com/hyperledger/aries-framework-go/blob/main/docs/vc_wallet.md) wallet.

This SDK provides wallet APIs for
* [Universal Wallet](https://w3c-ccg.github.io/universal-wallet-interop-spec/)
* [DIDComm](https://github.com/hyperledger/aries-rfcs/tree/master/concepts/0005-didcomm)
* [DIDComm Mediators](https://trustbloc.readthedocs.io/en/latest/agents/msg-routing-storage.html)
* [Blinded Routing](https://trustbloc.readthedocs.io/en/latest/blinded_routing.html)
* [DID management](https://github.com/hyperledger/aries-framework-go/blob/main/docs/concepts/00_what_is_hl_aries.md#9-vdr)
* [Device Registration](https://www.w3.org/TR/webauthn-2/)

A fully functional wallet can be built using the APIs provided in this wallet SDK.
For example: [TrustBloc User Agent Wallet](https://github.com/trustbloc/wallet) is built completely using this wallet SDK.

Refer TrustBloc User Agent wallet [documentation](https://github.com/trustbloc/wallet/blob/main/docs/components/web_wallet.md) for more details.

#### Data Models
Refer  Wallet SDK Data Model [documentation](data_models.md) to know about data models supported.

## API Reference
## Modules

<dl>
<dt><a href="#module_collection">collection</a></dt>
<dd><p>collection module provides wallet collection data model features for grouping wallet contents.
 This is useful for implementing credential vaults.</p>
</dd>
<dt><a href="#module_credential">credential</a></dt>
<dd><p>credential module provides wallet credential handling features,</p>
</dd>
<dt><a href="#module_device-login">device-login</a></dt>
<dd><p>device module provides device ogin features based on WebAuthN.</p>
</dd>
<dt><a href="#module_device-register">device-register</a></dt>
<dd><p>device module provides device registration features based on WebAuthN.</p>
</dd>
<dt><a href="#module_did-manager">did-manager</a></dt>
<dd><p>did-manager module provides DID related features for wallet like creating, importing &amp; saving DIDs into wallets.</p>
</dd>
<dt><a href="#module_blinded-routing">blinded-routing</a></dt>
<dd><p>blinded-routing module provides features supporting blinded DIDComm routing features.</p>
</dd>
<dt><a href="#module_didcomm">didcomm</a></dt>
<dd><p>didcomm module provides wallet based DIDComm features.</p>
</dd>
<dt><a href="#module_client">client</a></dt>
<dd><p>client module is the Authentication Server client that requests GNAP tokens from the Authorization Server.</p>
</dd>
<dt><a href="#module_HTTPSigner">HTTPSigner</a></dt>
<dd><p>HTTPSigner provides functionality to generate http signatures for requests</p>
</dd>
<dt><a href="#module_vcwallet">vcwallet</a></dt>
<dd><p>vcwallet module provides verifiable credential wallet SDK for aries universal wallet implementation.</p>
</dd>
<dt><a href="#module_wallet-user">wallet-user</a></dt>
<dd><p>wallet-user module provides wallet user specific features like maintaining profiles, preferences, locking and unlocking wallets.</p>
</dd>
</dl>

## Constants

<dl>
<dt><a href="#normalizePresentationSubmission">normalizePresentationSubmission</a> ⇒ <code>Array.&lt;Object&gt;</code></dt>
<dd><p>Scans through @see <a href="https://identity.foundation/presentation-exchange/#presentation-submission">Presentation Submission</a> descriptor map and groups results by descriptor IDs [descriptor_id -&gt; Array of verifiable Credentials].
 In many cases, a single input descriptor can find multiple credentials.
 So this function is useful in cases of grouping results by input descriptor ID and giving option to user to choose only one to avoid over sharing.</p>
<p> TODO: support for path_nested in descriptor map.</p>
</dd>
<dt><a href="#updatePresentationSubmission">updatePresentationSubmission</a></dt>
<dd><p>Updates given presentation submission presentation by removing duplicate descriptor map entries.</p>
<p> Descriptor map might contain single input descriptor ID mapped to multiple credentials.
 So returning PresentationSubmission presentation will retain only mappings mentioned in updates Object{<inputDescriptorID>:<credentialID>} parameter.</p>
</dd>
<dt><a href="#findAttachmentByFormat">findAttachmentByFormat</a></dt>
<dd><p>Finds attachment by given format.
 Supporting Attachment Format from DIDComm V1.</p>
<p> Note: Currently finding only one attachment per format.</p>
</dd>
<dt><a href="#findAttachmentByFormatV2">findAttachmentByFormatV2</a></dt>
<dd><p>Finds attachment by given format.
 Supporting Attachment Format from DIDComm V2.</p>
<p> Note: Currently finding only one attachment per format.</p>
</dd>
<dt><a href="#extractOOBGoalCode">extractOOBGoalCode</a></dt>
<dd><p>Reads out-of-band invitation goal code.
 Supports DIDComm V1 &amp; V2</p>
</dd>
</dl>

## Functions

<dl>
<dt><a href="#waitFor">waitFor()</a></dt>
<dd><p>Wait for given duration in millisecond and return promise.</p>
</dd>
</dl>

<a name="module_collection"></a>

## collection
collection module provides wallet collection data model features for grouping wallet contents.
 This is useful for implementing credential vaults.


* [collection](#module_collection)
    * [.exports.CollectionManager](#exp_module_collection--exports.CollectionManager) ⏏
        * [new exports.CollectionManager(agent, user)](#new_module_collection--exports.CollectionManager_new)
        * [.create(auth, collection)](#module_collection--exports.CollectionManager.CollectionManager+create) ⇒ <code>Promise.&lt;string&gt;</code>
        * [.getAll(auth)](#module_collection--exports.CollectionManager.CollectionManager+getAll) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.get(auth, collectionID)](#module_collection--exports.CollectionManager.CollectionManager+get) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.remove(auth, collectionID)](#module_collection--exports.CollectionManager.CollectionManager+remove) ⇒ <code>Promise</code>
        * [.update(auth, collectionID, collection)](#module_collection--exports.CollectionManager.CollectionManager+update) ⇒ <code>Promise</code>

<a name="exp_module_collection--exports.CollectionManager"></a>

### .exports.CollectionManager ⏏
Creating, updating, deleting, querying collections

**Kind**: static class of [<code>collection</code>](#module_collection)  
<a name="new_module_collection--exports.CollectionManager_new"></a>

#### new exports.CollectionManager(agent, user)

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>string</code> | aries agent. |
| user | <code>string</code> | unique wallet user identifier, the one used to create wallet profile. |

<a name="module_collection--exports.CollectionManager.CollectionManager+create"></a>

#### exports.CollectionManager.create(auth, collection) ⇒ <code>Promise.&lt;string&gt;</code>
Creates new wallet collection model and adds it to wallet content store.

**Kind**: instance method of [<code>exports.CollectionManager</code>](#exp_module_collection--exports.CollectionManager)  
**Returns**: <code>Promise.&lt;string&gt;</code> - - promise containing collection ID or an error if operation fails..  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| auth | <code>String</code> |  | authorization token for wallet operations. |
| collection | <code>Object</code> |  | collection data model. |
| collection.name | <code>String</code> |  | display name of the collection. |
| collection.description | <code>String</code> |  | display description of the collection. |
| collection.type | <code>String</code> | <code>ContentVault</code> | optional, custom collection type, default='ContentVault'.  TODO: support for more customized collection parameters like icon, color, storage params etc  for supporting different types of credential vaults. |

<a name="module_collection--exports.CollectionManager.CollectionManager+getAll"></a>

#### exports.CollectionManager.getAll(auth) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets list of all collections from wallet content store.

**Kind**: instance method of [<code>exports.CollectionManager</code>](#exp_module_collection--exports.CollectionManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing collection contents or an error if operation fails..  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |

<a name="module_collection--exports.CollectionManager.CollectionManager+get"></a>

#### exports.CollectionManager.get(auth, collectionID) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets a collection from wallet content store.

**Kind**: instance method of [<code>exports.CollectionManager</code>](#exp_module_collection--exports.CollectionManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing collection content or an error if operation fails..  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| collectionID | <code>String</code> | ID of the collection to retrieved from store. |

<a name="module_collection--exports.CollectionManager.CollectionManager+remove"></a>

#### exports.CollectionManager.remove(auth, collectionID) ⇒ <code>Promise</code>
Removes a collection from wallet content store and also deletes all contents which belongs to the collection.

**Kind**: instance method of [<code>exports.CollectionManager</code>](#exp_module_collection--exports.CollectionManager)  
**Returns**: <code>Promise</code> - - empty promise or an error if operation fails..  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| collectionID | <code>String</code> | ID of the collection to retrieved from store. |

<a name="module_collection--exports.CollectionManager.CollectionManager+update"></a>

#### exports.CollectionManager.update(auth, collectionID, collection) ⇒ <code>Promise</code>
Updates a collection from wallet content store.

**Kind**: instance method of [<code>exports.CollectionManager</code>](#exp_module_collection--exports.CollectionManager)  
**Returns**: <code>Promise</code> - - empty promise or an error if operation fails..  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| collectionID | <code>String</code> | ID of the collection to retrieved from store. |
| collection | <code>Object</code> | collection data model. |
| collection.name | <code>String</code> | display name of the collection. |
| collection.description | <code>String</code> | display description of the collection. |

<a name="module_credential"></a>

## credential
credential module provides wallet credential handling features,


* [credential](#module_credential)
    * [.exports.CredentialManager](#exp_module_credential--exports.CredentialManager) ⏏
        * [new exports.CredentialManager(agent, user)](#new_module_credential--exports.CredentialManager_new)
        * [.save(auth, contents, options)](#module_credential--exports.CredentialManager.CredentialManager+save) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.saveCredentialMetadata(auth, options)](#module_credential--exports.CredentialManager.CredentialManager+saveCredentialMetadata) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.getCredentialMetadata(auth, id)](#module_credential--exports.CredentialManager.CredentialManager+getCredentialMetadata) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.getAllCredentialMetadata(auth, options)](#module_credential--exports.CredentialManager.CredentialManager+getAllCredentialMetadata) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.updateCredentialMetadata(auth, id, options)](#module_credential--exports.CredentialManager.CredentialManager+updateCredentialMetadata) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.resolveManifest(auth, options)](#module_credential--exports.CredentialManager.CredentialManager+resolveManifest) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.get(auth, contentID)](#module_credential--exports.CredentialManager.CredentialManager+get) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.getAll(auth)](#module_credential--exports.CredentialManager.CredentialManager+getAll) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.remove(auth, contentID)](#module_credential--exports.CredentialManager.CredentialManager+remove) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.issue(auth, credential, proofOptions)](#module_credential--exports.CredentialManager.CredentialManager+issue) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.present(auth, credentialOptions, proofOptions)](#module_credential--exports.CredentialManager.CredentialManager+present) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.verify(auth, verificationOption)](#module_credential--exports.CredentialManager.CredentialManager+verify) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.derive(auth, credentialOption, deriveOption)](#module_credential--exports.CredentialManager.CredentialManager+derive) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.query(auth, query)](#module_credential--exports.CredentialManager.CredentialManager+query) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.saveManifestVC(auth, manifest, connectionID)](#module_credential--exports.CredentialManager.CredentialManager+saveManifestVC) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.getManifestConnection(auth, manifestCredID)](#module_credential--exports.CredentialManager.CredentialManager+getManifestConnection) ⇒ <code>Promise.&lt;String&gt;</code>
        * [.getAllManifestVCs(auth)](#module_credential--exports.CredentialManager.CredentialManager+getAllManifestVCs) ⇒ <code>Promise.&lt;Object&gt;</code>

<a name="exp_module_credential--exports.CredentialManager"></a>

### .exports.CredentialManager ⏏
- Save, get, remove & get all.
 - Issue, prove, verify & derive.
 - query (QueryByExample, QueryByFrame, PresentationExchange & DIDAuth)

**Kind**: static class of [<code>credential</code>](#module_credential)  
<a name="new_module_credential--exports.CredentialManager_new"></a>

#### new exports.CredentialManager(agent, user)

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>string</code> | aries agent. |
| user | <code>string</code> | unique wallet user identifier, the one used to create wallet profile. |

<a name="module_credential--exports.CredentialManager.CredentialManager+save"></a>

#### exports.CredentialManager.save(auth, contents, options) ⇒ <code>Promise.&lt;Object&gt;</code>
Saves given credential into wallet content store along with credential metadata & manifest details along with saved credential.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>string</code> | authorization token for wallet operations. |
| contents | <code>Object</code> | credential(s) to be saved in wallet content store. |
| contents.credentials | <code>Array.&lt;Object&gt;</code> | array of credentials to be saved in wallet content store. |
| contents.presentation | <code>Object</code> | presentation from which all the credentials to be saved in wallet content store.  If credential response presentation is provided then no need to supply descriptor map along with manifest.  Refer @see [Credential Response](https://identity.foundation/credential-manifest/#credential-response) for more details. |
| options | <code>Object</code> | options for saving credential. |
| options.verify | <code>boolean</code> | (optional) to verify credential before save. |
| options.collection | <code>String</code> | (optional) ID of the wallet collection to which the credential should belong. |
| options.manifest | <code>String</code> | (required) credential manifest of the credential being saved.  Refer @see [Credential Manifest](https://identity.foundation/credential-manifest/#credential-manifest-2) for more details. |

<a name="module_credential--exports.CredentialManager.CredentialManager+saveCredentialMetadata"></a>

#### exports.CredentialManager.saveCredentialMetadata(auth, options) ⇒ <code>Promise.&lt;Object&gt;</code>
Reads credential metadata and saves credential metadata data model into wallet content store.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>string</code> | authorization token for wallet operations. |
| options | <code>Object</code> | subjects from which credential metadata will be extracted. |
| options.credential | <code>Object</code> | credential data model from which basic credential attributes like type, issuer, expiration etc will be read. |
| options.manifestID | <code>String</code> | ID of the credential manifest of the given credential |
| options.descriptorID | <code>String</code> | ID of the credential manifest output descriptor of the given credential. |
| options.collection | <code>String</code> | (optional) ID of the collection to which this credential belongs. |

<a name="module_credential--exports.CredentialManager.CredentialManager+getCredentialMetadata"></a>

#### exports.CredentialManager.getCredentialMetadata(auth, id) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets credential metadata from wallet content store and also optionally resolves credential data using credential manifest.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing credential metadata or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| id | <code>String</code> | credential ID. |

<a name="module_credential--exports.CredentialManager.CredentialManager+getAllCredentialMetadata"></a>

#### exports.CredentialManager.getAllCredentialMetadata(auth, options) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets all credential metadata models from wallet content store.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing list of credential metadata or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| options | <code>Object</code> | options to get all credential metadata. |
| options.credentialIDs | <code>Bool</code> | (optional) filters credential metadata by given credential IDs. |
| options.collection | <code>String</code> | (optional) filters credential metadata by given collection ID. |

<a name="module_credential--exports.CredentialManager.CredentialManager+updateCredentialMetadata"></a>

#### exports.CredentialManager.updateCredentialMetadata(auth, id, options) ⇒ <code>Promise.&lt;Object&gt;</code>
Updates credential metadata. Currently supporting updating only credential name and description fields.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| id | <code>String</code> | ID of the credential metadata to be updated. |
| options | <code>Object</code> | options to update credential metadata. |
| options.name | <code>String</code> | (optional) name attribute of the credential metadata to be updated. |
| options.description | <code>String</code> | (optional) description attribute of the credential metadata to be updated. |
| options.collection | <code>String</code> | (optional) ID  of the collection to which this credential metadata to be updated. |

<a name="module_credential--exports.CredentialManager.CredentialManager+resolveManifest"></a>

#### exports.CredentialManager.resolveManifest(auth, options) ⇒ <code>Promise.&lt;Object&gt;</code>
Resolves credential by credential manifest, descriptor or response.

Given credential can be resolved by raw credential, ID of the credential saved in wallet, credential response,
ID of the manifest saved in wallet, raw credential manifest, output descriptor of the manifest etc

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing resolved results or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| options | <code>Object</code> | options to resolve credential from wallet. |
| options.credentialID | <code>String</code> | (optional) ID of the credential to be resolved from wallet content store. |
| options.credential | <code>String</code> | (optional) raw credential data model to be resolved. |
| options.response | <code>String</code> | (optional) credential response using which given raw credential or credential ID to be resolved. |
| options.manifestID | <code>String</code> | (optional) ID of the manifest from wallet content store. |
| options.manifest | <code>String</code> | (optional) raw manifest to be used for resolving credential. |
| options.descriptorID | <code>String</code> | (optional) if response not provided then this descriptor ID can be used to resolve credential. Refer @see [Credential Manifest Specifications](https://identity.foundation/credential-manifest/) for more details. |

<a name="module_credential--exports.CredentialManager.CredentialManager+get"></a>

#### exports.CredentialManager.get(auth, contentID) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets credential from wallet

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - result.content -- promise containing credential or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>string</code> | authorization token for wallet operations. |
| contentID | <code>string</code> | ID of the credential to be read from wallet content store. |

<a name="module_credential--exports.CredentialManager.CredentialManager+getAll"></a>

#### exports.CredentialManager.getAll(auth) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets All credentials from wallet.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - result.contents - promise containing results or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>string</code> | authorization token for wallet operations. |

<a name="module_credential--exports.CredentialManager.CredentialManager+remove"></a>

#### exports.CredentialManager.remove(auth, contentID) ⇒ <code>Promise.&lt;Object&gt;</code>
Removes credential and its metadata from wallet.

 Doesn't delete respective credential manifest since one credential manifest can be referred by many other credentials too.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>string</code> | authorization token for wallet operations. |
| contentID | <code>string</code> | ID of the credential to be removed from wallet content store. |

<a name="module_credential--exports.CredentialManager.CredentialManager+issue"></a>

#### exports.CredentialManager.issue(auth, credential, proofOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
Issues credential from wallet

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing issued credential or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>string</code> | authorization token for wallet operations. |
| credential | <code>Object</code> | credential to be signed from wallet. |
| proofOptions | <code>Object</code> | credential to be signed from wallet. |
| proofOptions.controller | <code>string</code> | DID to be used for signing. |
| proofOptions.verificationMethod | <code>string</code> | (optional) VerificationMethod is the URI of the verificationMethod used for the proof.  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions. |
| proofOptions.created | <code>string</code> | (optional) Created date of the proof.  By default, current system time will be used. |
| proofOptions.domain | <code>string</code> | (optional) operational domain of a digital proof.  By default, domain will not be part of proof. |
| proofOptions.challenge | <code>string</code> | (optional) random or pseudo-random value option authentication.  By default, challenge will not be part of proof. |
| proofOptions.proofType | <code>string</code> | (optional) signature type used for signing.  By default, proof will be generated in Ed25519Signature2018 format. |
| proofOptions.proofFormat | <code>String</code> | (optional) representational format for the credential.  Valid values are "ExternalJWTProofFormat" and "EmbeddedLDProofFormat".  By default, credential will be JSON-LD with embedded proof. |
| proofOptions.proofRepresentation | <code>string</code> | (optional) type of proof data expected ( "proofValue" or "jws").  By default, 'proofValue' will be used. |

<a name="module_credential--exports.CredentialManager.CredentialManager+present"></a>

#### exports.CredentialManager.present(auth, credentialOptions, proofOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
Prepares verifiable presentation of given credential(s).

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of signed presentation or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| credentialOptions | <code>Object</code> | credential/presentations to verify.. |
| credentialOptions.storedCredentials | <code>Array.&lt;string&gt;</code> | (optional) ids of the credentials already saved in wallet content store. |
| credentialOptions.rawCredentials | <code>Array.&lt;Object&gt;</code> | (optional) list of raw credentials to be presented. |
| credentialOptions.presentation | <code>Object</code> | (optional) presentation to be proved. |
| proofOptions | <code>Object</code> | proof options for signing presentation. |
| proofOptions.controller | <code>String</code> | DID to be used for signing. |
| proofOptions.verificationMethod | <code>String</code> | (optional) VerificationMethod is the URI of the verificationMethod used for the proof.  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions. |
| proofOptions.created | <code>String</code> | (optional) Created date of the proof.  By default, current system time will be used. |
| proofOptions.domain | <code>String</code> | (optional) operational domain of a digital proof.  By default, domain will not be part of proof. |
| proofOptions.challenge | <code>String</code> | (optional) random or pseudo-random value option authentication.  By default, challenge will not be part of proof. |
| proofOptions.proofType | <code>String</code> | (optional) signature type used for signing.  By default, proof will be generated in Ed25519Signature2018 format. |
| proofOptions.proofFormat | <code>String</code> | (optional) representational format for the presentation.  Valid values are "ExternalJWTProofFormat" and "EmbeddedLDProofFormat".  By default, presentation will be JSON-LD with embedded proof. |
| proofOptions.proofRepresentation | <code>String</code> | (optional) type of proof data expected ( "proofValue" or "jws").  By default, 'proofValue' will be used. |

<a name="module_credential--exports.CredentialManager.CredentialManager+verify"></a>

#### exports.CredentialManager.verify(auth, verificationOption) ⇒ <code>Promise.&lt;Object&gt;</code>
Verifies credential/presentation from wallet.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of verification result(bool) and error containing cause if verification fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| verificationOption | <code>String</code> | credential/presentation to be verified. |
| verificationOption.storedCredentialID | <code>String</code> | (optional) id of the credential already saved in wallet content store. |
| verificationOption.rawCredential | <code>Object</code> | (optional) credential to be verified. |
| verificationOption.presentation | <code>Object</code> | (optional) presentation to be verified. |

<a name="module_credential--exports.CredentialManager.CredentialManager+derive"></a>

#### exports.CredentialManager.derive(auth, credentialOption, deriveOption) ⇒ <code>Promise.&lt;Object&gt;</code>
Derives a credential from wallet.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of derived credential or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| credentialOption | <code>String</code> | credential to be dervied. |
| credentialOption.storedCredentialID | <code>String</code> | (optional) id of the credential already saved in wallet content store. |
| credentialOption.rawCredential | <code>Object</code> | (optional) credential to be derived. |
| deriveOption | <code>Object</code> | derive options. |
| deriveOption.frame | <code>Object</code> | JSON-LD frame used for derivation. |
| deriveOption.nonce | <code>String</code> | (optional) to prove uniqueness or freshness of the proof.. |

<a name="module_credential--exports.CredentialManager.CredentialManager+query"></a>

#### exports.CredentialManager.query(auth, query) ⇒ <code>Promise.&lt;Object&gt;</code>
runs credential queries in wallet.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of presentation(s) containing credential results or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| query | <code>Object</code> | list of credential queries, any types of supported query types can be mixed. |

<a name="module_credential--exports.CredentialManager.CredentialManager+saveManifestVC"></a>

#### exports.CredentialManager.saveManifestVC(auth, manifest, connectionID) ⇒ <code>Promise.&lt;Object&gt;</code>
saves manifest credential along with its mapping to given connection ID.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or an error if operation fails.  
**Deprecated,**: to be used for DIDComm blinded routing flow only  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| manifest | <code>Object</code> | manifest credential (can be of any type). |
| connectionID | <code>String</code> | connection ID to which manifest credential to be mapped. |

<a name="module_credential--exports.CredentialManager.CredentialManager+getManifestConnection"></a>

#### exports.CredentialManager.getManifestConnection(auth, manifestCredID) ⇒ <code>Promise.&lt;String&gt;</code>
Returns connection ID mapped to given manifest credential ID.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;String&gt;</code> - - promise containing connection ID or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| manifestCredID | <code>String</code> | ID of manifest credential. |

<a name="module_credential--exports.CredentialManager.CredentialManager+getAllManifestVCs"></a>

#### exports.CredentialManager.getAllManifestVCs(auth) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets all manifest credentials saved in wallet.

**Kind**: instance method of [<code>exports.CredentialManager</code>](#exp_module_credential--exports.CredentialManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing manifest credential search results or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |

<a name="module_device-login"></a>

## device-login
device module provides device ogin features based on WebAuthN.


* [device-login](#module_device-login)
    * [.exports.DeviceLogin](#exp_module_device-login--exports.DeviceLogin) ⏏
        * [new exports.DeviceLogin(serverURL)](#new_module_device-login--exports.DeviceLogin_new)
        * [.login()](#module_device-login--exports.DeviceLogin.DeviceLogin+login)

<a name="exp_module_device-login--exports.DeviceLogin"></a>

### .exports.DeviceLogin ⏏
DeviceLogin provides device login features.

**Kind**: static class of [<code>device-login</code>](#module_device-login)  
<a name="new_module_device-login--exports.DeviceLogin_new"></a>

#### new exports.DeviceLogin(serverURL)

| Param | Type | Description |
| --- | --- | --- |
| serverURL | <code>string</code> | device login server URL. |

<a name="module_device-login--exports.DeviceLogin.DeviceLogin+login"></a>

#### exports.DeviceLogin.login()
Performs Device Login.

**Kind**: instance method of [<code>exports.DeviceLogin</code>](#exp_module_device-login--exports.DeviceLogin)  
<a name="module_device-register"></a>

## device-register
device module provides device registration features based on WebAuthN.


* [device-register](#module_device-register)
    * [.exports.DeviceRegister](#exp_module_device-register--exports.DeviceRegister) ⏏
        * [new exports.DeviceRegister(serverURL)](#new_module_device-register--exports.DeviceRegister_new)
        * [.register()](#module_device-register--exports.DeviceRegister.DeviceRegister+register)

<a name="exp_module_device-register--exports.DeviceRegister"></a>

### .exports.DeviceRegister ⏏
DeviceRegister provides device registration features.

**Kind**: static class of [<code>device-register</code>](#module_device-register)  
<a name="new_module_device-register--exports.DeviceRegister_new"></a>

#### new exports.DeviceRegister(serverURL)

| Param | Type | Description |
| --- | --- | --- |
| serverURL | <code>string</code> | device login server URL. |

<a name="module_device-register--exports.DeviceRegister.DeviceRegister+register"></a>

#### exports.DeviceRegister.register()
Performs Device Registration.

**Kind**: instance method of [<code>exports.DeviceRegister</code>](#exp_module_device-register--exports.DeviceRegister)  
<a name="module_did-manager"></a>

## did-manager
did-manager module provides DID related features for wallet like creating, importing & saving DIDs into wallets.


* [did-manager](#module_did-manager)
    * [.exports.DIDManager](#exp_module_did-manager--exports.DIDManager) ⏏
        * [new exports.DIDManager(agent, user)](#new_module_did-manager--exports.DIDManager_new)
        * [.createOrbDID(auth, options)](#module_did-manager--exports.DIDManager.DIDManager+createOrbDID) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.createPeerDID(options)](#module_did-manager--exports.DIDManager.DIDManager+createPeerDID) ⇒ <code>Promise</code>
        * [.saveDID(options)](#module_did-manager--exports.DIDManager.DIDManager+saveDID) ⇒ <code>Promise</code>
        * [.importDID(options)](#module_did-manager--exports.DIDManager.DIDManager+importDID) ⇒ <code>Promise</code>
        * [.getAllDIDs(options)](#module_did-manager--exports.DIDManager.DIDManager+getAllDIDs) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.getDID(options)](#module_did-manager--exports.DIDManager.DIDManager+getDID) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.resolveOrbDID(options)](#module_did-manager--exports.DIDManager.DIDManager+resolveOrbDID) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.resolveWebDIDFromOrbDID(options)](#module_did-manager--exports.DIDManager.DIDManager+resolveWebDIDFromOrbDID) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.verifyWebDIDFromOrbDID(options)](#module_did-manager--exports.DIDManager.DIDManager+verifyWebDIDFromOrbDID) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.checkControllerIsPublished(options)](#module_did-manager--exports.DIDManager.DIDManager+checkControllerIsPublished) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.removeDID(options)](#module_did-manager--exports.DIDManager.DIDManager+removeDID) ⇒ <code>Promise.&lt;Object&gt;</code>

<a name="exp_module_did-manager--exports.DIDManager"></a>

### .exports.DIDManager ⏏
DID Manger provides DID related features for wallet like,

 - Creating Orb DIDs.
 - Creating Peer DIDs.
 - Saving Custom DIDs along with keys.
 - Getting all Saved DIDs.

**Kind**: static class of [<code>did-manager</code>](#module_did-manager)  
<a name="new_module_did-manager--exports.DIDManager_new"></a>

#### new exports.DIDManager(agent, user)

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>string</code> | aries agent. |
| user | <code>string</code> | unique wallet user identifier, the one used to create wallet profile. |

<a name="module_did-manager--exports.DIDManager.DIDManager+createOrbDID"></a>

#### exports.DIDManager.createOrbDID(auth, options) ⇒ <code>Promise.&lt;Object&gt;</code>
Creates Orb DID and saves it in wallet content store.

If DID is not anchored (equivalentId ID found in DID document metadata) then saves DID resolved from equivalent ID.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - Promise of DID Resolution response  or an error if operation fails..  
**See**: [The did:orb Method](https://trustbloc.github.io/did-method-orb)  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| auth | <code>string</code> |  | authorization token for wallet operations. |
| options | <code>Object</code> |  | options for creating Orb DID. |
| options.keyType | <code>Object</code> | <code>ED25519</code> | (optional, default ED25519) type of the key to be used for creating keys for the DID, Refer agent documentation for supported key types. |
| options.keyAgreementKeyType | <code>Object</code> | <code>X25519ECDHKW</code> | (optional, default X25519ECDHKW) type of the key to be used for creating keyAgreements for the DID, Refer agent documentation for supported key types. |
| options.signatureType | <code>String</code> | <code>Ed25519VerificationKey2018</code> | (optional, default Ed25519VerificationKey2018) signature type to be used for DID verification methods. |
| options.keyAgreementType | <code>String</code> | <code>X25519KeyAgreementKey2019</code> | (optional, default X25519KeyAgreementKey2019) keyAgreement VM type to be used for DID key agreement (payload encryption). For JWK type, use `JsonWebKey2020`. |
| options.purposes | <code>Array.&lt;String&gt;</code> | <code>authentication</code> | (optional, default "authentication") purpose of the key. |
| options.routerKeyAgreementIDs | <code>Array.&lt;String&gt;</code> | <code>[</code> | (optional, used for DIDComm V2 only, default empty list) list of router keys IDs. |
| options.routerConnections | <code>Array.&lt;String&gt;</code> | <code>[</code> | (optional, used for DIDComm V2 only, default empty list) list of router connections. |
| options.serviceID | <code>String</code> |  | (optional, default no serviceID set) serviceID to which this DID should belong to. |
| options.serviceEndpoint | <code>String</code> |  | (optional, default no serviceEndpoint set) serviceEndpoint to which this DID should have its service accessible. |
| options.didcommServiceType | <code>String</code> |  | (optional, default no didcommServiceType set) didcommServiceType to which this DID belong to (didcomm v1: "did-communication", or didcomm V2: "DIDCommMessaging"). |
| options.collection | <code>String</code> |  | (optional, default no collection) collection to which this DID should belong in wallet content store. |

<a name="module_did-manager--exports.DIDManager.DIDManager+createPeerDID"></a>

#### exports.DIDManager.createPeerDID(options) ⇒ <code>Promise</code>
Creates Peer DID and saves it in wallet content store.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise</code> - - empty promise or an error if operation fails..  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.collection | <code>string</code> | (optional, default no collection) collection to which this DID should belong in wallet content store. |

<a name="module_did-manager--exports.DIDManager.DIDManager+saveDID"></a>

#### exports.DIDManager.saveDID(options) ⇒ <code>Promise</code>
Saves given DID content to wallet content store.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise</code> - - empty promise or an error if operation fails..  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.content | <code>string</code> | DID document content. |
| options.collection | <code>string</code> | (optional, default no collection) collection to which this DID should belong in wallet content store. |

<a name="module_did-manager--exports.DIDManager.DIDManager+importDID"></a>

#### exports.DIDManager.importDID(options) ⇒ <code>Promise</code>
Resolves and saves DID document into wallet content store along with keys.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise</code> - - empty promise or an error if operation fails..  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.did | <code>string</code> | ID of the DID to be imported. |
| options.key | <code>string</code> | (optional, default no collection) collection to which this DID should belong in wallet content store. |
| options.collection | <code>string</code> | (optional, default no collection) collection to which this DID should belong in wallet content store. |

<a name="module_did-manager--exports.DIDManager.DIDManager+getAllDIDs"></a>

#### exports.DIDManager.getAllDIDs(options) ⇒ <code>Promise.&lt;Object&gt;</code>
gets all DID contents from wallet content store.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - result.contents - collection of DID documents by IDs.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.collection | <code>string</code> | (optional, default no collection) to filter DID contents based on collection ID. |

<a name="module_did-manager--exports.DIDManager.DIDManager+getDID"></a>

#### exports.DIDManager.getDID(options) ⇒ <code>Promise.&lt;Object&gt;</code>
get DID content from wallet content store.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - result.content - DID document resolution from wallet content store.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.contentID | <code>string</code> | DID ID. |

<a name="module_did-manager--exports.DIDManager.DIDManager+resolveOrbDID"></a>

#### exports.DIDManager.resolveOrbDID(options) ⇒ <code>Promise.&lt;Object&gt;</code>
resolve orb DID.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - result.content - DID document resolution from did resolver.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.contentID | <code>string</code> | DID ID. |

<a name="module_did-manager--exports.DIDManager.DIDManager+resolveWebDIDFromOrbDID"></a>

#### exports.DIDManager.resolveWebDIDFromOrbDID(options) ⇒ <code>Promise.&lt;Object&gt;</code>
resolve web DID from orb DID.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - result.content - DID document resolution from did resolver.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.contentID | <code>string</code> | DID ID. |

<a name="module_did-manager--exports.DIDManager.DIDManager+verifyWebDIDFromOrbDID"></a>

#### exports.DIDManager.verifyWebDIDFromOrbDID(options) ⇒ <code>Promise.&lt;Object&gt;</code>
verify web DID from orb DID.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - result.content - DID document resolution from did resolver.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.contentID | <code>string</code> | DID ID. |

<a name="module_did-manager--exports.DIDManager.DIDManager+checkControllerIsPublished"></a>

#### exports.DIDManager.checkControllerIsPublished(options) ⇒ <code>Promise.&lt;Object&gt;</code>
check controller if it is published.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - true if controller is published.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.contentID | <code>string</code> | DID ID (typically orb https domain ID). |

<a name="module_did-manager--exports.DIDManager.DIDManager+removeDID"></a>

#### exports.DIDManager.removeDID(options) ⇒ <code>Promise.&lt;Object&gt;</code>
removes given DID from wallet content store.

**Kind**: instance method of [<code>exports.DIDManager</code>](#exp_module_did-manager--exports.DIDManager)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.auth | <code>string</code> | authorization token for wallet operations. |
| options.contentID | <code>string</code> | DID ID of the DID to be deleted. |

<a name="module_blinded-routing"></a>

## blinded-routing
blinded-routing module provides features supporting blinded DIDComm routing features.


* [blinded-routing](#module_blinded-routing)
    * [.exports.BlindedRouter](#exp_module_blinded-routing--exports.BlindedRouter) ⏏
        * [new exports.BlindedRouter(agent)](#new_module_blinded-routing--exports.BlindedRouter_new)
        * [.sharePeerDID(connection)](#module_blinded-routing--exports.BlindedRouter.BlindedRouter+sharePeerDID)

<a name="exp_module_blinded-routing--exports.BlindedRouter"></a>

### .exports.BlindedRouter ⏏
BlindedRouter provides DIDComm message based blinded routing features.

**Kind**: static class of [<code>blinded-routing</code>](#module_blinded-routing)  
<a name="new_module_blinded-routing--exports.BlindedRouter_new"></a>

#### new exports.BlindedRouter(agent)

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>string</code> | aries agent. |

<a name="module_blinded-routing--exports.BlindedRouter.BlindedRouter+sharePeerDID"></a>

#### exports.BlindedRouter.sharePeerDID(connection)
This function provides functionality of sharing peer DID with connecting party for blinded DIDComm.

**Kind**: instance method of [<code>exports.BlindedRouter</code>](#exp_module_blinded-routing--exports.BlindedRouter)  

| Param | Type | Description |
| --- | --- | --- |
| connection | <code>Object</code> | connection record of the connection established. |

<a name="module_didcomm"></a>

## didcomm
didcomm module provides wallet based DIDComm features.


* [didcomm](#module_didcomm)
    * [.exports.DIDComm](#exp_module_didcomm--exports.DIDComm) ⏏
        * [new exports.DIDComm(agent, user)](#new_module_didcomm--exports.DIDComm_new)
        * _instance_
            * [.connect(auth, invitation, options)](#module_didcomm--exports.DIDComm.DIDComm+connect) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.initiateCredentialShare(auth, invitation, connectOptions, proposeOptions)](#module_didcomm--exports.DIDComm.DIDComm+initiateCredentialShare) ⇒ <code>Object</code> \| <code>String</code> \| <code>Array.&lt;Object&gt;</code> \| <code>Array.&lt;Object&gt;</code>
            * [.completeCredentialShare(auth, threadID, presentations, proofOptions, options)](#module_didcomm--exports.DIDComm.DIDComm+completeCredentialShare) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.initiateCredentialIssuance(auth, invitation, connectOptions, proposeOptions)](#module_didcomm--exports.DIDComm.DIDComm+initiateCredentialIssuance) ⇒ <code>Object</code> \| <code>String</code> \| <code>Object</code> \| <code>Object</code> \| <code>Object</code> \| <code>Array.&lt;Object&gt;</code> \| <code>Array.&lt;Object&gt;</code> \| <code>String</code> \| <code>String</code> \| <code>String</code>
            * [.completeCredentialIssuance(auth, threadID, presentation, proofOptions, options)](#module_didcomm--exports.DIDComm.DIDComm+completeCredentialIssuance) ⇒ <code>Promise.&lt;Object&gt;</code>
        * _static_
            * [.createInvitationFromRouter](#module_didcomm--exports.DIDComm.createInvitationFromRouter)
            * [.getMediatorConnections(agent)](#module_didcomm--exports.DIDComm.getMediatorConnections)
            * [.connectToMediator(agent, endpoint, waitForStateComplete, isDIDCommV2)](#module_didcomm--exports.DIDComm.connectToMediator)
        * _inner_
            * [~getPresentationAttachmentAndThreadID(presentationRequest)](#module_didcomm--exports.DIDComm..getPresentationAttachmentAndThreadID)

<a name="exp_module_didcomm--exports.DIDComm"></a>

### .exports.DIDComm ⏏
DIDComm module provides wallet based DIDComm features. Currently supporting DID-Exchange, Present Proof & WACI features.

**Kind**: static class of [<code>didcomm</code>](#module_didcomm)  
<a name="new_module_didcomm--exports.DIDComm_new"></a>

#### new exports.DIDComm(agent, user)

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>string</code> | aries agent. |
| user | <code>string</code> | unique wallet user identifier, the one used to create wallet profile. |

<a name="module_didcomm--exports.DIDComm.DIDComm+connect"></a>

#### exports.DIDComm.connect(auth, invitation, options) ⇒ <code>Promise.&lt;Object&gt;</code>
accepts an out of band invitation, performs did-exchange and returns connection record of connection established.

**Kind**: instance method of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing connection ID or error if operation fails.  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| auth | <code>String</code> |  | authorization token for performing this wallet operation. |
| invitation | <code>Object</code> |  | out of band invitation. |
| options | <code>Object</code> |  | (optional) for accepting incoming out-of-band invitation and connecting to inviter. |
| options.myLabel | <code>String</code> |  | (optional) for providing label to be shared with the other agent during the subsequent did-exchange. |
| options.routerConnections | <code>Array.&lt;string&gt;</code> |  | (optional) to provide router connection to be used. |
| options.userAnyRouterConnection | <code>Bool</code> | <code>false</code> | (optional) if true and options.routerConnections not provided then wallet will find  an existing router connection and will use it for accepting invitation. |
| options.reuseConnection | <code>String</code> |  | (optional) to provide DID to be used when reusing a connection. |
| options.reuseAnyConnection | <code>Bool</code> | <code>false</code> | (optional) to use any recognized DID in the services array for a reusable connection. |
| options.timeout | <code>Time</code> |  | (optional) to wait for connection status to be 'completed'. |
| options.waitForCompletion | <code>Bool</code> |  | (optional) if true then wait for custom 'didexchange-state-complete' message to conclude connection as completed. |

<a name="module_didcomm--exports.DIDComm.DIDComm+initiateCredentialShare"></a>

#### exports.DIDComm.initiateCredentialShare(auth, invitation, connectOptions, proposeOptions) ⇒ <code>Object</code> \| <code>String</code> \| <code>Array.&lt;Object&gt;</code> \| <code>Array.&lt;Object&gt;</code>
Initiates WACI credential share interaction from wallet.

 accepts an out of band invitation, sends propose presentation message to inviter, waits for request presentation message reply from inviter.
 reads presentation definition(s) from request presentation, performs query in wallet and returns response presentation(s) to be shared.

**Kind**: instance method of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  
**Returns**: <code>Object</code> - response - promise of object containing presentation request message from relying party or error if operation fails.<code>String</code> - response.threadID - thread ID of credential interaction to be used for correlation.<code>Array.&lt;Object&gt;</code> - response.presentations - array of presentation responses from wallet query.<code>Array.&lt;Object&gt;</code> - response.normalized - normalized version of `response.presentations` where response credentials are grouped by input descriptors.
Can be used to detect multiple credential result for same query.  
**See**: [WACI Presentation flow ](https://identity.foundation/waci-presentation-exchange/#presentation-2) for more details.  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| auth | <code>String</code> |  | authorization token for performing this wallet operation. |
| invitation | <code>Object</code> |  | out of band invitation. |
| connectOptions | <code>Object</code> |  | (optional) for accepting incoming out-of-band invitation and connecting to inviter. |
| connectOptions.myLabel | <code>String</code> |  | (optional) for providing label to be shared with the other agent during the subsequent did-exchange. |
| connectOptions.routerConnections | <code>Array.&lt;string&gt;</code> |  | (optional) to provide router connection to be used. |
| options.userAnyRouterConnection | <code>Bool</code> | <code>false</code> | (optional) if true and options.routerConnections not provided then wallet will find  an existing router connection and will use it for accepting invitation. |
| connectOptions.reuseConnection | <code>String</code> |  | (optional) to provide DID to be used when reusing a connection. |
| connectOptions.reuseAnyConnection | <code>Bool</code> | <code>false</code> | (optional) to use any recognized DID in the services array for a reusable connection. |
| connectOptions.connectionTimeout | <code>timeout</code> |  | (optional) to wait for connection status to be 'completed'. |
| proposeOptions | <code>Object</code> |  | (optional) for sending message proposing presentation. |
| proposeOptions.from | <code>String</code> |  | (optional) option from DID option to customize sender DID.. |
| proposeOptions.timeout | <code>Time</code> |  | (optional) to wait for request presentation message from relying party. |

<a name="module_didcomm--exports.DIDComm.DIDComm+completeCredentialShare"></a>

#### exports.DIDComm.completeCredentialShare(auth, threadID, presentations, proofOptions, options) ⇒ <code>Promise.&lt;Object&gt;</code>
Completes WACI credential share flow.

 Signs presentation(s) and sends them as part of present proof message to relying party.

**Kind**: instance method of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing present prof status & redirect info or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| threadID | <code>String</code> | threadID of credential interaction. |
| presentations | <code>Array.&lt;Object&gt;</code> | to be sent as part of present proof message.. |
| proofOptions | <code>Object</code> | proof options for signing presentation. |
| proofOptions.controller | <code>String</code> | DID to be used for signing. |
| proofOptions.verificationMethod | <code>String</code> | (optional) VerificationMethod is the URI of the verificationMethod used for the proof.  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions. |
| proofOptions.created | <code>String</code> | (optional) Created date of the proof.  By default, current system time will be used. |
| proofOptions.domain | <code>String</code> | (optional) operational domain of a digital proof.  By default, domain will not be part of proof. |
| proofOptions.challenge | <code>String</code> | (optional) random or pseudo-random value option authentication.  By default, challenge will not be part of proof. |
| proofOptions.proofType | <code>String</code> | (optional) signature type used for signing.  By default, proof will be generated in Ed25519Signature2018 format. |
| proofOptions.proofRepresentation | <code>String</code> | (optional) type of proof data expected ( "proofValue" or "jws").  By default, 'proofValue' will be used. |
| options | <code>Object</code> | (optional) for sending message proposing presentation. |
| options.waitForDone | <code>Bool</code> | (optional) If true then wallet will wait for present proof protocol status to be done or abandoned . |
| options.WaitForDoneTimeout | <code>Time</code> | (optional) timeout to wait for present proof operation to be done. |
| options.autoAccept | <code>Bool</code> | (optional) can be used to auto accept any incoming problem reports while waiting for present proof protocol status to be done or abandoned. |

<a name="module_didcomm--exports.DIDComm.DIDComm+initiateCredentialIssuance"></a>

#### exports.DIDComm.initiateCredentialIssuance(auth, invitation, connectOptions, proposeOptions) ⇒ <code>Object</code> \| <code>String</code> \| <code>Object</code> \| <code>Object</code> \| <code>Object</code> \| <code>Array.&lt;Object&gt;</code> \| <code>Array.&lt;Object&gt;</code> \| <code>String</code> \| <code>String</code> \| <code>String</code>
Initiates WACI credential issuance interaction from wallet.

 accepts an out of band invitation, sends request credential message to inviter, waits for offer credential message response from inviter.

 If present, reads presentation definition(s) from offer credential message, performs query in wallet and returns response presentation(s) to be shared.

**Kind**: instance method of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  
**Returns**: <code>Object</code> - response - promise of object containing offer credential message from issuer or error if operation fails.<code>String</code> - response.threadID - thread ID of credential interaction, to be used for correlation in future.<code>Object</code> - response.error - error containing status, code and redirect URL if requested by issuer.<code>Object</code> - manifest - credential manifest sent by issuer.<code>Object</code> - response - credential response sent by issuer.<code>Array.&lt;Object&gt;</code> - response.presentations - array of presentation responses from wallet query.<code>Array.&lt;Object&gt;</code> - response.normalized - normalized version of `response.presentations` where response credentials are grouped by input descriptors.<code>String</code> - domain - domain parameter sent by issuer for proving ownership of DID or freshness of proof.<code>String</code> - challenge - challenge parameter sent by issuer for proving ownership of DID or freshness of proof..<code>String</code> - comment - custom comment sent by issuer along with credential response.
Can be used to detect multiple credential result for same query.  
**See**: [WACI Issuance flow ](https://identity.foundation/waci-presentation-exchange/#issuance-2) for more details.  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| auth | <code>String</code> |  | authorization token for performing this wallet operation. |
| invitation | <code>Object</code> |  | out of band invitation. |
| connectOptions | <code>Object</code> |  | (optional) for accepting incoming out-of-band invitation and connecting to inviter. |
| connectOptions.myLabel | <code>String</code> |  | (optional) for providing label to be shared with the other agent during the subsequent did-exchange. |
| connectOptions.routerConnections | <code>Array.&lt;string&gt;</code> |  | (optional) to provide router connection to be used. |
| options.userAnyRouterConnection | <code>Bool</code> | <code>false</code> | (optional) if true and options.routerConnections not provided then wallet will find  an existing router connection and will use it for accepting invitation. |
| connectOptions.reuseConnection | <code>String</code> |  | (optional) to provide DID to be used when reusing a connection. |
| connectOptions.reuseAnyConnection | <code>Bool</code> | <code>false</code> | (optional) to use any recognized DID in the services array for a reusable connection. |
| connectOptions.connectionTimeout | <code>timeout</code> |  | (optional) to wait for connection status to be 'completed'. |
| proposeOptions | <code>Object</code> |  | (optional) for sending message proposing credential. |
| proposeOptions.from | <code>String</code> |  | (optional) option from DID option to customize sender DID.. |
| proposeOptions.timeout | <code>Time</code> |  | (optional) to wait for offer credential message from relying party. |

<a name="module_didcomm--exports.DIDComm.DIDComm+completeCredentialIssuance"></a>

#### exports.DIDComm.completeCredentialIssuance(auth, threadID, presentation, proofOptions, options) ⇒ <code>Promise.&lt;Object&gt;</code>
Completes WACI credential issuance flow.

 Sends request credential message to issuer as part of ongoing WACI issuance flow and waits for credential response response from issuer.
 Optionally sends presentations as credential application attachments as part of request credential message.

 Response credentials from credential response will be saved to collection of choice.

**Kind**: instance method of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing request credential status & redirect info or error if operation fails.  
**See**: [WACI Issuance flow ](https://identity.foundation/waci-presentation-exchange/#issuance-2) for more details.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| threadID | <code>String</code> | threadID of credential interaction. |
| presentation | <code>Object</code> | to be sent as part of credential response. This presentations will be converted into credential response format |
| proofOptions | <code>Object</code> | proof options for signing presentation. |
| proofOptions.controller | <code>String</code> | DID to be used for signing. |
| proofOptions.verificationMethod | <code>String</code> | (optional) VerificationMethod is the URI of the verificationMethod used for the proof.  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions. |
| proofOptions.created | <code>String</code> | (optional) Created date of the proof.  By default, current system time will be used. |
| proofOptions.domain | <code>String</code> | (optional) operational domain of a digital proof.  By default, domain will not be part of proof. |
| proofOptions.challenge | <code>String</code> | (optional) random or pseudo-random value option authentication.  By default, challenge will not be part of proof. |
| proofOptions.proofType | <code>String</code> | (optional) signature type used for signing.  By default, proof will be generated in Ed25519Signature2018 format. |
| proofOptions.proofRepresentation | <code>String</code> | (optional) type of proof data expected ( "proofValue" or "jws").  By default, 'proofValue' will be used. |
| options | <code>Object</code> | (optional) for sending message requesting credential. |
| options.waitForDone | <code>Bool</code> | (optional) If true then wallet will wait for credential response message or problem report to arrive. |
| options.WaitForDoneTimeout | <code>Time</code> | (optional) timeout to wait for for credential response message or problem report to arrive. Will be considered only  when `options.waitForDone` is true. |
| options.autoAccept | <code>Bool</code> | (optional) if enabled then incoming issue credential or problem report will be auto accepted. If not provided then  wallet will rely on underlying agent to accept incoming actions. |
| options.collection | <code>String</code> | (optional) ID of the wallet collection to which the credential should belong. |

<a name="module_didcomm--exports.DIDComm.createInvitationFromRouter"></a>

#### exports.DIDComm.createInvitationFromRouter
Get DID Invitation from edge router.

**Kind**: static constant of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  

| Param | Description |
| --- | --- |
| endpoint | edge router endpoint |
| isDIDCommV2 | flag using DIDComm V2 |

<a name="module_didcomm--exports.DIDComm.getMediatorConnections"></a>

#### exports.DIDComm.getMediatorConnections(agent)
Get router/mediator connections from agent.

**Kind**: static method of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  

| Param | Description |
| --- | --- |
| agent | instance |

<a name="module_didcomm--exports.DIDComm.connectToMediator"></a>

#### exports.DIDComm.connectToMediator(agent, endpoint, waitForStateComplete, isDIDCommV2)
Connect given agent to edge mediator/router.

**Kind**: static method of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  

| Param | Description |
| --- | --- |
| agent | trustbloc agent |
| endpoint | edge router endpoint |
| waitForStateComplete | wait for did exchange state complete message |
| isDIDCommV2 | flag using DIDComm V2 |

<a name="module_didcomm--exports.DIDComm..getPresentationAttachmentAndThreadID"></a>

#### exports.DIDComm~getPresentationAttachmentAndThreadID(presentationRequest)
Get attachment and threadID from presentationRequest instance based on DIDComm V1 or V2 formats.

**Kind**: inner method of [<code>exports.DIDComm</code>](#exp_module_didcomm--exports.DIDComm)  

| Param | Description |
| --- | --- |
| presentationRequest | instance |

<a name="module_client"></a>

## client
client module is the Authentication Server client that requests GNAP tokens from the Authorization Server.

<a name="exp_module_client--exports.Client"></a>

### .exports.Client ⏏
Client requesting Gnap tokens from the Authorization Server.

**Kind**: static class of [<code>client</code>](#module_client)  
<a name="module_HTTPSigner"></a>

## HTTPSigner
HTTPSigner provides functionality to generate http signatures for requests


* [HTTPSigner](#module_HTTPSigner)
    * [.exports.HTTPSigner](#exp_module_HTTPSigner--exports.HTTPSigner) ⏏
        * [.generateSignatureParams()](#module_HTTPSigner--exports.HTTPSigner.HTTPSigner+generateSignatureParams) ⇒ <code>String</code>
        * [.getSignatureInput(name, sigParams)](#module_HTTPSigner--exports.HTTPSigner.HTTPSigner+getSignatureInput) ⇒ <code>String</code>
        * [.proofType()](#module_HTTPSigner--exports.HTTPSigner.HTTPSigner+proofType) ⇒ <code>String</code>
        * [.sign(digest, url, name, sigParams)](#module_HTTPSigner--exports.HTTPSigner.HTTPSigner+sign) ⇒ <code>String</code>

<a name="exp_module_HTTPSigner--exports.HTTPSigner"></a>

### .exports.HTTPSigner ⏏
- generate signature params.
 - get signature input.
 - sign requests.

**Kind**: static class of [<code>HTTPSigner</code>](#module_HTTPSigner)  
<a name="module_HTTPSigner--exports.HTTPSigner.HTTPSigner+generateSignatureParams"></a>

#### exports.HTTPSigner.generateSignatureParams() ⇒ <code>String</code>
Generates and returns signature params string.

**Kind**: instance method of [<code>exports.HTTPSigner</code>](#exp_module_HTTPSigner--exports.HTTPSigner)  
**Returns**: <code>String</code> - - generated signature params  
<a name="module_HTTPSigner--exports.HTTPSigner.HTTPSigner+getSignatureInput"></a>

#### exports.HTTPSigner.getSignatureInput(name, sigParams) ⇒ <code>String</code>
Generates and returns a signature input for the name and signature params provided.

**Kind**: instance method of [<code>exports.HTTPSigner</code>](#exp_module_HTTPSigner--exports.HTTPSigner)  
**Returns**: <code>String</code> - - generated signature input  

| Param | Type | Description |
| --- | --- | --- |
| name | <code>String</code> | name of the signature |
| sigParams | <code>String</code> | signature parameters |

<a name="module_HTTPSigner--exports.HTTPSigner.HTTPSigner+proofType"></a>

#### exports.HTTPSigner.proofType() ⇒ <code>String</code>
Returns the proof type of the http signature.

**Kind**: instance method of [<code>exports.HTTPSigner</code>](#exp_module_HTTPSigner--exports.HTTPSigner)  
**Returns**: <code>String</code> - - proof type of the http signature  
<a name="module_HTTPSigner--exports.HTTPSigner.HTTPSigner+sign"></a>

#### exports.HTTPSigner.sign(digest, url, name, sigParams) ⇒ <code>String</code>
Generates and returns http signature for the request provided.

**Kind**: instance method of [<code>exports.HTTPSigner</code>](#exp_module_HTTPSigner--exports.HTTPSigner)  
**Returns**: <code>String</code> - - generated signature for the data provided  

| Param | Type | Description |
| --- | --- | --- |
| digest | <code>String</code> | digest value for the request that's being signed |
| url | <code>String</code> | url of the server |
| name | <code>String</code> | signature name |
| sigParams | <code>String</code> | signature parameters |

<a name="module_vcwallet"></a>

## vcwallet
vcwallet module provides verifiable credential wallet SDK for aries universal wallet implementation.


* [vcwallet](#module_vcwallet)
    * [.exports.UniversalWallet](#exp_module_vcwallet--exports.UniversalWallet) ⏏
        * [new exports.UniversalWallet(agent, user)](#new_module_vcwallet--exports.UniversalWallet_new)
        * _instance_
            * [.open(options)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+open) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.close()](#module_vcwallet--exports.UniversalWallet.UniversalWallet+close) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.add(request)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+add) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.remove(request)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+remove) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.get(request)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+get) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.getAll(request)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+getAll) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.query(auth, query)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+query) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.issue(auth, credential, proofOptions)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+issue) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.prove(auth, credentialOptions, proofOptions)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+prove) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.verify(auth, verificationOption)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+verify) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.derive(auth, credentialOption, deriveOption)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+derive) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.createKeyPair(request)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+createKeyPair) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.connect(auth, invitation, options)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+connect) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.proposePresentation(auth, invitation, connectOptions, proposeOptions)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+proposePresentation) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.presentProof(auth, threadID, presentation, options)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+presentProof) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.proposeCredential(auth, invitation, connectOptions, proposeOptions)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+proposeCredential) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.requestCredential(auth, threadID, presentation, options)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+requestCredential) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.resolveCredential(auth, manifest, options)](#module_vcwallet--exports.UniversalWallet.UniversalWallet+resolveCredential) ⇒ <code>Promise.&lt;Object&gt;</code>
        * _static_
            * [.contentTypes](#module_vcwallet--exports.UniversalWallet.contentTypes) : <code>enum</code>
            * [.createWalletProfile(agent, userID, profileOptions)](#module_vcwallet--exports.UniversalWallet.createWalletProfile) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.updateWalletProfile(agent, userID, profileOptions)](#module_vcwallet--exports.UniversalWallet.updateWalletProfile) ⇒ <code>Promise.&lt;Object&gt;</code>
            * [.profileExists(agent, userID, profilestorage)](#module_vcwallet--exports.UniversalWallet.profileExists) ⇒ <code>Promise.&lt;Object&gt;</code>

<a name="exp_module_vcwallet--exports.UniversalWallet"></a>

### .exports.UniversalWallet ⏏
UniversalWallet is universal wallet SDK built on top aries verifiable credential wallet controller (vcwallet).

https://w3c-ccg.github.io/universal-wallet-interop-spec/

Aries JS Controller: https://github.com/hyperledger/aries-framework-go/blob/main/docs/vc_wallet.md#javascript

Refer Agent SDK Open API spec for detailed vcwallet request response models.

**Kind**: static class of [<code>vcwallet</code>](#module_vcwallet)  
<a name="new_module_vcwallet--exports.UniversalWallet_new"></a>

#### new exports.UniversalWallet(agent, user)

| Param | Description |
| --- | --- |
| agent | aries agent. |
| user | unique wallet user identifier, the one used to create wallet profile. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+open"></a>

#### exports.UniversalWallet.open(options) ⇒ <code>Promise.&lt;Object&gt;</code>
Unlocks given wallet's key manager instance & content store and
returns a authorization token to be used for performing wallet operations.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - 'object.token' - auth token subsequent use of wallet features.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.localKMSPassphrase | <code>String</code> | (optional) passphrase for local kms for key operations. |
| options.webKMSAuth | <code>Object</code> | (optional) WebKMSAuth for authorizing access to web/remote kms. |
| options.webKMSAuth.authToken | <code>String</code> | (optional) Http header 'authorization' bearer token to be used, i.e access token. |
| options.webKMSAuth.capability | <code>String</code> | (optional) Capability if ZCAP sign header feature to be used for authorizing access. |
| options.webKMSAuth.authzKeyStoreURL | <code>String</code> | (optional) authz key store URL if ZCAP sign header feature to be used for authorizing access. |
| options.webKMSAuth.secretShare | <code>String</code> | (optional) secret share if ZCAP sign header feature to be used for authorizing access. |
| options.edvUnlocks | <code>Object</code> | (optional) for authorizing access to wallet's EDV content store. |
| options.edvUnlocks.authToken | <code>String</code> | (optional) Http header 'authorization' bearer token to be used, i.e access token. |
| options.edvUnlocks.capability | <code>String</code> | (optional) Capability if ZCAP sign header feature to be used for authorizing access. |
| options.edvUnlocks.authzKeyStoreURL | <code>String</code> | (optional) authz key store URL if ZCAP sign header feature to be used for authorizing access. |
| options.edvUnlocks.secretShare | <code>String</code> | (optional) secret share if ZCAP sign header feature to be used for authorizing access. |
| options.expiry | <code>Time</code> | (optional) time duration in milliseconds for which this profile will be unlocked. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+close"></a>

#### exports.UniversalWallet.close() ⇒ <code>Promise.&lt;Object&gt;</code>
Expires token issued to this VC wallet, removes wallet's key manager instance and closes wallet content store.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - 'object.closed' -  bool flag false if token is not found or already expired for this wallet user.  
<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+add"></a>

#### exports.UniversalWallet.add(request) ⇒ <code>Promise.&lt;Object&gt;</code>
Adds given content to wallet content store.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or an error if adding content to wallet store fails.  

| Param | Type | Description |
| --- | --- | --- |
| request | <code>Object</code> |  |
| request.auth | <code>String</code> | authorization token for performing this wallet operation. |
| request.contentType | <code>Object</code> | type of the content to be added to the wallet, refer aries vc wallet for supported types. |
| request.content | <code>String</code> | content to be added wallet store. |
| request.collectionID | <code>String</code> | (optional) ID of the wallet collection to which the content should belong. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+remove"></a>

#### exports.UniversalWallet.remove(request) ⇒ <code>Promise.&lt;Object&gt;</code>
remove given content from wallet content store.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| request | <code>Object</code> |  |
| request.auth | <code>String</code> | authorization token for performing this wallet operation. |
| request.contentType | <code>Object</code> | type of the content to be removed from the wallet. |
| request.contentID | <code>String</code> | id of the content to be removed from wallet. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+get"></a>

#### exports.UniversalWallet.get(request) ⇒ <code>Promise.&lt;Object&gt;</code>
gets wallet content by ID from wallet content store.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing content or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| request | <code>Object</code> |  |
| request.auth | <code>String</code> | authorization token for performing this wallet operation. |
| request.contentType | <code>Object</code> | type of the content to be removed from the wallet. |
| request.contentID | <code>String</code> | id of the content to be returned from wallet. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+getAll"></a>

#### exports.UniversalWallet.getAll(request) ⇒ <code>Promise.&lt;Object&gt;</code>
gets all wallet contents from wallet content store for given type.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing response contents or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| request | <code>Object</code> |  |
| request.auth | <code>String</code> | authorization token for performing this wallet operation. |
| request.contentType | <code>Object</code> | type of the contents to be returned from wallet. |
| request.collectionID | <code>String</code> | id of the collection on which the response contents to be filtered. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+query"></a>

#### exports.UniversalWallet.query(auth, query) ⇒ <code>Promise.&lt;Object&gt;</code>
runs credential queries against wallet credential contents.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of presentation(s) containing credential results or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| query | <code>Object</code> | credential query, refer: https://w3c-ccg.github.io/vp-request-spec/#format |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+issue"></a>

#### exports.UniversalWallet.issue(auth, credential, proofOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
runs credential queries against wallet credential contents.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of credential issued or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| credential | <code>Object</code> | credential to be signed from wallet. |
| proofOptions | <code>Object</code> | proof options for issuing credential. |
| proofOptions.controller | <code>String</code> | DID to be used for signing. |
| proofOptions.verificationMethod | <code>String</code> | (optional) VerificationMethod is the URI of the verificationMethod used for the proof.  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions. |
| proofOptions.created | <code>String</code> | (optional) Created date of the proof.  By default, current system time will be used. |
| proofOptions.domain | <code>String</code> | (optional) operational domain of a digital proof.  By default, domain will not be part of proof. |
| proofOptions.challenge | <code>String</code> | (optional) random or pseudo-random value option authentication.  By default, challenge will not be part of proof. |
| proofOptions.proofType | <code>String</code> | (optional) signature type used for signing.  By default, proof will be generated in Ed25519Signature2018 format. |
| proofOptions.proofRepresentation | <code>String</code> | (optional) type of proof data expected ( "proofValue" or "jws").  By default, 'proofValue' will be used. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+prove"></a>

#### exports.UniversalWallet.prove(auth, credentialOptions, proofOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
produces a Verifiable Presentation from wallet.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of signed presentation or an error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| credentialOptions | <code>Object</code> | credential/presentations to verify.. |
| credentialOptions.storedCredentials | <code>Array.&lt;string&gt;</code> | (optional) ids of the credentials already saved in wallet content store. |
| credentialOptions.rawCredentials | <code>Array.&lt;Object&gt;</code> | (optional) list of raw credentials to be presented. |
| credentialOptions.presentation | <code>Object</code> | (optional) presentation to be proved. |
| proofOptions | <code>Object</code> | proof options for signing. |
| proofOptions.controller | <code>String</code> | DID to be used for signing. |
| proofOptions.verificationMethod | <code>String</code> | (optional) VerificationMethod is the URI of the verificationMethod used for the proof.  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions. |
| proofOptions.created | <code>String</code> | (optional) Created date of the proof.  By default, current system time will be used. |
| proofOptions.domain | <code>String</code> | (optional) operational domain of a digital proof.  By default, domain will not be part of proof. |
| proofOptions.challenge | <code>String</code> | (optional) random or pseudo-random value option authentication.  By default, challenge will not be part of proof. |
| proofOptions.proofType | <code>String</code> | (optional) signature type used for signing.  By default, proof will be generated in Ed25519Signature2018 format. |
| proofOptions.proofRepresentation | <code>String</code> | (optional) type of proof data expected ( "proofValue" or "jws").  By default, 'proofValue' will be used. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+verify"></a>

#### exports.UniversalWallet.verify(auth, verificationOption) ⇒ <code>Promise.&lt;Object&gt;</code>
verifies credential/presentation from wallet.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of verification result(bool) and error containing cause if verification fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| verificationOption | <code>String</code> | credential/presentation to be verified. |
| verificationOption.storedCredentialID | <code>String</code> | (optional) id of the credential already saved in wallet content store. |
| verificationOption.rawCredential | <code>Object</code> | (optional) credential to be verified. |
| verificationOption.presentation | <code>Object</code> | (optional) presentation to be verified. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+derive"></a>

#### exports.UniversalWallet.derive(auth, credentialOption, deriveOption) ⇒ <code>Promise.&lt;Object&gt;</code>
derives a credential from wallet.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of derived credential or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| credentialOption | <code>String</code> | credential to be dervied. |
| credentialOption.storedCredentialID | <code>String</code> | (optional) id of the credential already saved in wallet content store. |
| credentialOption.rawCredential | <code>Object</code> | (optional) credential to be derived. |
| deriveOption | <code>Object</code> | derive options. |
| deriveOption.frame | <code>Object</code> | JSON-LD frame used for derivation. |
| deriveOption.nonce | <code>String</code> | (optional) to prove uniqueness or freshness of the proof.. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+createKeyPair"></a>

#### exports.UniversalWallet.createKeyPair(request) ⇒ <code>Promise.&lt;Object&gt;</code>
creates a key pair from wallet.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of derived credential or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| request | <code>Object</code> |  |
| request.auth | <code>String</code> | authorization token for performing this wallet operation. |
| request.keyType | <code>String</code> | type of the key to be created, refer aries kms for supported key types. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+connect"></a>

#### exports.UniversalWallet.connect(auth, invitation, options) ⇒ <code>Promise.&lt;Object&gt;</code>
accepts an out of band invitation and performs did-exchange.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing connection ID or error if operation fails.  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| auth | <code>String</code> |  | authorization token for performing this wallet operation. |
| invitation | <code>Object</code> |  | out of band invitation. |
| options | <code>Object</code> |  | (optional) for accepting incoming out-of-band invitation and connecting to inviter. |
| options.myLabel | <code>String</code> |  | (optional) for providing label to be shared with the other agent during the subsequent did-exchange. |
| options.routerConnections | <code>Array.&lt;string&gt;</code> |  | (optional) to provide router connection to be used. |
| options.reuseConnection | <code>String</code> |  | (optional) to provide DID to be used when reusing a connection. |
| options.reuseAnyConnection | <code>Bool</code> | <code>false</code> | (optional) to use any recognized DID in the services array for a reusable connection. |
| options.timeout | <code>Time</code> |  | (optional) to wait for connection status to be 'completed'. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+proposePresentation"></a>

#### exports.UniversalWallet.proposePresentation(auth, invitation, connectOptions, proposeOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
accepts an out of band invitation, sends propose presentation message to inviter to initiate credential share interaction
 and waits for request presentation message from inviter as a response.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing presentation request message from relying party or error if operation fails.  
**See**: [WACI Propose Presentation ](https://w3c-ccg.github.io/universal-wallet-interop-spec/#proposepresentation)  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| auth | <code>String</code> |  | authorization token for performing this wallet operation. |
| invitation | <code>Object</code> |  | out of band invitation. |
| connectOptions | <code>Object</code> |  | (optional) for accepting incoming out-of-band invitation and connecting to inviter. |
| connectOptions.myLabel | <code>String</code> |  | (optional) for providing label to be shared with the other agent during the subsequent did-exchange. |
| connectOptions.routerConnections | <code>Array.&lt;string&gt;</code> |  | (optional) to provide router connection to be used. |
| connectOptions.reuseConnection | <code>String</code> |  | (optional) to provide DID to be used when reusing a connection. |
| connectOptions.reuseAnyConnection | <code>Bool</code> | <code>false</code> | (optional) to use any recognized DID in the services array for a reusable connection. |
| connectOptions.connectionTimeout | <code>timeout</code> |  | (optional) to wait for connection status to be 'completed'. |
| proposeOptions | <code>Object</code> |  | (optional) for sending message proposing presentation. |
| proposeOptions.from | <code>String</code> |  | (optional) option from DID option to customize sender DID.. |
| proposeOptions.timeout | <code>Time</code> |  | (optional) to wait for request presentation message from relying party. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+presentProof"></a>

#### exports.UniversalWallet.presentProof(auth, threadID, presentation, options) ⇒ <code>Promise.&lt;Object&gt;</code>
sends present proof message from wallet to relying party as part of ongoing credential share interaction.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing present proof status & redirect info or error if operation fails.  
**See**: [WACI Present Proof ](https://w3c-ccg.github.io/universal-wallet-interop-spec/#presentproof)  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| threadID | <code>String</code> | threadID of credential interaction. |
| presentation | <code>Object</code> | to be sent as part of present proof message.. |
| options | <code>Object</code> | (optional) for sending present proof message. |
| options.waitForDone | <code>Bool</code> | (optional) If true then wallet will wait for present proof protocol status to be done or abandoned . |
| options.WaitForDoneTimeout | <code>Time</code> | (optional) timeout to wait for present proof operation to be done. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+proposeCredential"></a>

#### exports.UniversalWallet.proposeCredential(auth, invitation, connectOptions, proposeOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
accepts an out of band invitation, sends propose credential message to issuer to initiate credential issuance interaction
 and waits for offer credential message from inviter as a response.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing offer credential message from relying party or error if operation fails.  
**See**: [WACI Propose Credential ](https://w3c-ccg.github.io/universal-wallet-interop-spec/#proposecredential)  

| Param | Type | Default | Description |
| --- | --- | --- | --- |
| auth | <code>String</code> |  | authorization token for performing this wallet operation. |
| invitation | <code>Object</code> |  | out of band invitation. |
| connectOptions | <code>Object</code> |  | (optional) for accepting incoming out-of-band invitation and connecting to inviter. |
| connectOptions.myLabel | <code>String</code> |  | (optional) for providing label to be shared with the other agent during the subsequent did-exchange. |
| connectOptions.routerConnections | <code>Array.&lt;string&gt;</code> |  | (optional) to provide router connection to be used. |
| connectOptions.reuseConnection | <code>String</code> |  | (optional) to provide DID to be used when reusing a connection. |
| connectOptions.reuseAnyConnection | <code>Bool</code> | <code>false</code> | (optional) to use any recognized DID in the services array for a reusable connection. |
| connectOptions.connectionTimeout | <code>timeout</code> |  | (optional) to wait for connection status to be 'completed'. |
| proposeOptions | <code>Object</code> |  | (optional) for sending message proposing credential. |
| proposeOptions.from | <code>String</code> |  | (optional) option from DID option to customize sender DID.. |
| proposeOptions.timeout | <code>Time</code> |  | (optional) to wait for offer credential message from relying party. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+requestCredential"></a>

#### exports.UniversalWallet.requestCredential(auth, threadID, presentation, options) ⇒ <code>Promise.&lt;Object&gt;</code>
sends request credential message from wallet to issuer as part of ongoing credential issuance interaction.

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing request credential status & redirect info or error if operation fails.  
**See**: [WACI Request Credential ](https://w3c-ccg.github.io/universal-wallet-interop-spec/#requestcredential)  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| threadID | <code>String</code> | threadID of credential interaction. |
| presentation | <code>Object</code> | to be sent as part of request credential message.. |
| options | <code>Object</code> | (optional) for sending request credential message. |
| options.waitForDone | <code>Bool</code> | (optional) If true then wallet will wait for credential response message or problem report . |
| options.WaitForDoneTimeout | <code>Time</code> | (optional) timeout to wait for credential response or problem report. |

<a name="module_vcwallet--exports.UniversalWallet.UniversalWallet+resolveCredential"></a>

#### exports.UniversalWallet.resolveCredential(auth, manifest, options) ⇒ <code>Promise.&lt;Object&gt;</code>
resolves given credential manifest by credential response or credential.
Supports: https://identity.foundation/credential-manifest/

**Kind**: instance method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise of object containing request credential status & redirect info or error if operation fails.  
**See**: [Credential Manifest ](https://identity.foundation/credential-manifest)  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for performing this wallet operation. |
| manifest | <code>Object</code> | credential manifest to be used for resolving credential. |
| options | <code>Object</code> | credential or response to resolve. |
| options.response | <code>String</code> | (optional) credential response to be resolved.  if provided, then this option takes precedence over credential resolve option. |
| options.credential | <code>Object</code> | (optional) raw credential to be resolved (accepting 'ldp_vc' format only).  This option has to be provided with descriptor ID. |
| options.credentialID | <code>String</code> | (optional) ID of the credential to be resolved which is persisted in wallet content store.  This option has to be provided with descriptor ID. |
| options.descriptorID | <code>String</code> | (optional) output descriptor ID of the descriptor from manifest to be used for resolving given  credential or credentialID. This option is required only when a raw credential or credential ID is to be resolved. |

<a name="module_vcwallet--exports.UniversalWallet.contentTypes"></a>

#### exports.UniversalWallet.contentTypes : <code>enum</code>
Supported content type from this wallet.

**Kind**: static enum of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**See**: [Aries VC Wallet Data Models](https://github.com/hyperledger/aries-framework-go/blob/main/docs/vc_wallet.md#supported-data-models)  
<a name="module_vcwallet--exports.UniversalWallet.createWalletProfile"></a>

#### exports.UniversalWallet.createWalletProfile(agent, userID, profileOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
creates new wallet profile for given user.

**Kind**: static method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>Object</code> | aries agent |
| userID | <code>String</code> | unique identifier of user for which the profile is being created. |
| profileOptions | <code>String</code> | options for creating profile. |
| profileOptions.localKMSPassphrase | <code>String</code> | (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations. |
| profileOptions.keyStoreURL | <code>String</code> | (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations. |
| profileOptions.edvConfiguration | <code>Object</code> | (optional) EDV configuration if profile wants to use EDV as a wallet content store.  By Default, aries context storage provider will be used. |
| profileOptions.edvConfiguration.serverURL | <code>String</code> | EDV server URL for storing wallet contents. |
| profileOptions.edvConfiguration.vaultID | <code>String</code> | EDV vault ID for storing the wallet contents. |
| profileOptions.edvConfiguration.encryptionKID | <code>String</code> | Encryption key ID of already existing key in wallet profile kms.  If profile is using localkms then wallet will create this key set for wallet user. |
| profileOptions.edvConfiguration.macKID | <code>String</code> | MAC operation key ID of already existing key in wallet profile kms.  If profile is using localkms then wallet will create this key set for wallet user. |

<a name="module_vcwallet--exports.UniversalWallet.updateWalletProfile"></a>

#### exports.UniversalWallet.updateWalletProfile(agent, userID, profileOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
updates existing wallet profile for given user.
 Caution:
 - you might lose your existing keys if you change kms options.
 - you might lose your existing wallet contents if you change storage/EDV options
 (ex: switching context storage provider or changing EDV settings).

**Kind**: static method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>Object</code> | aries agent |
| userID | <code>String</code> | unique identifier of user for which the profile is being created. |
| profileOptions | <code>String</code> | options for creating profile. |
| profileOptions.localKMSPassphrase | <code>String</code> | (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations. |
| profileOptions.keyStoreURL | <code>String</code> | (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations. |
| profileOptions.edvConfiguration | <code>String</code> | (optional) EDV configuration if profile wants to use EDV as a wallet content store.  By Default, aries context storage provider will be used. |

<a name="module_vcwallet--exports.UniversalWallet.profileExists"></a>

#### exports.UniversalWallet.profileExists(agent, userID, profilestorage) ⇒ <code>Promise.&lt;Object&gt;</code>
check is profile exists for given wallet user.

**Kind**: static method of [<code>exports.UniversalWallet</code>](#exp_module_vcwallet--exports.UniversalWallet)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if profile not found.  

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>Object</code> | aries agent |
| userID | <code>String</code> | unique identifier of user for which the profile is being created. |
| profilestorage | <code>String</code> | provider will be used. |

<a name="module_wallet-user"></a>

## wallet-user
wallet-user module provides wallet user specific features like maintaining profiles, preferences, locking and unlocking wallets.


* [wallet-user](#module_wallet-user)
    * [.exports.WalletUser](#exp_module_wallet-user--exports.WalletUser) ⏏
        * [new exports.WalletUser(agent, user)](#new_module_wallet-user--exports.WalletUser_new)
        * [.createWalletProfile(profileOptions)](#module_wallet-user--exports.WalletUser.WalletUser+createWalletProfile) ⇒ <code>Promise</code>
        * [.updateWalletProfile(profileOptions)](#module_wallet-user--exports.WalletUser.WalletUser+updateWalletProfile) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.profileExists()](#module_wallet-user--exports.WalletUser.WalletUser+profileExists) ⇒ <code>Promise.&lt;Boolean&gt;</code>
        * [.unlock(options)](#module_wallet-user--exports.WalletUser.WalletUser+unlock) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.lock()](#module_wallet-user--exports.WalletUser.WalletUser+lock) ⇒ <code>Promise.&lt;Bool&gt;</code>
        * [.savePreferences(auth, preferences)](#module_wallet-user--exports.WalletUser.WalletUser+savePreferences) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.updatePreferences(auth, preferences)](#module_wallet-user--exports.WalletUser.WalletUser+updatePreferences) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.getPreferences(auth)](#module_wallet-user--exports.WalletUser.WalletUser+getPreferences) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.saveMetadata(auth, content)](#module_wallet-user--exports.WalletUser.WalletUser+saveMetadata) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.getMetadata(auth, contentID)](#module_wallet-user--exports.WalletUser.WalletUser+getMetadata) ⇒ <code>Promise.&lt;Object&gt;</code>
        * [.getAllMetadata(auth)](#module_wallet-user--exports.WalletUser.WalletUser+getAllMetadata) ⇒ <code>Promise.&lt;Object&gt;</code>

<a name="exp_module_wallet-user--exports.WalletUser"></a>

### .exports.WalletUser ⏏
WalletUser provides wallet user related features like,

 - Creating and updating wallet user profiles.
 - Saving and updating user wallet preferences.
 - Unlocking and locking wallet.

**Kind**: static class of [<code>wallet-user</code>](#module_wallet-user)  
<a name="new_module_wallet-user--exports.WalletUser_new"></a>

#### new exports.WalletUser(agent, user)

| Param | Type | Description |
| --- | --- | --- |
| agent | <code>String</code> | aries agent. |
| user | <code>String</code> | unique wallet user identifier, the one used to create wallet profile. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+createWalletProfile"></a>

#### exports.WalletUser.createWalletProfile(profileOptions) ⇒ <code>Promise</code>
Create wallet profile for the user and returns error if profile is already created.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise</code> - - empty promise or an error if operation fails..  

| Param | Type | Description |
| --- | --- | --- |
| profileOptions | <code>String</code> | options for creating profile. |
| profileOptions.localKMSPassphrase | <code>String</code> | (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations. |
| profileOptions.keyStoreURL | <code>String</code> | (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations. |
| profileOptions.edvConfiguration | <code>String</code> | (optional) EDV configuration if profile wants to use EDV as a wallet content store.  By Default, aries context storage provider will be used. |
| profileOptions.edvConfiguration.serverURL | <code>String</code> | EDV server URL for storing wallet contents. |
| profileOptions.edvConfiguration.vaultID | <code>String</code> | EDV vault ID for storing the wallet contents. |
| profileOptions.edvConfiguration.encryptionKID | <code>String</code> | Encryption key ID of already existing key in wallet profile kms.  If profile is using localkms then wallet will create this key set for wallet user. |
| profileOptions.edvConfiguration.macKID | <code>String</code> | MAC operation key ID of already existing key in wallet profile kms.  If profile is using localkms then wallet will create this key set for wallet user. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+updateWalletProfile"></a>

#### exports.WalletUser.updateWalletProfile(profileOptions) ⇒ <code>Promise.&lt;Object&gt;</code>
Updates wallet profile for the user and returns error if profile doesn't exists.
Caution:
 - you might lose your existing keys if you change kms options.
 - you might lose your existing wallet contents if you change storage/EDV options
 (ex: switching context storage provider or changing EDV settings).

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| profileOptions | <code>String</code> | options for creating profile. |
| profileOptions.localKMSPassphrase | <code>String</code> | (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations. |
| profileOptions.keyStoreURL | <code>String</code> | (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations. |
| profileOptions.edvConfiguration | <code>String</code> | (optional) EDV configuration if profile wants to use EDV as a wallet content store.  By Default, aries context storage provider will be used. |
| profileOptions.edvConfiguration.serverURL | <code>String</code> | EDV server URL for storing wallet contents. |
| profileOptions.edvConfiguration.vaultID | <code>String</code> | EDV vault ID for storing the wallet contents. |
| profileOptions.edvConfiguration.encryptionKID | <code>String</code> | Encryption key ID of already existing key in wallet profile kms.  If profile is using localkms then wallet will create this key set for wallet user. |
| profileOptions.edvConfiguration.macKID | <code>String</code> | MAC operation key ID of already existing key in wallet profile kms.  If profile is using localkms then wallet will create this key set for wallet user. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+profileExists"></a>

#### exports.WalletUser.profileExists() ⇒ <code>Promise.&lt;Boolean&gt;</code>
check is profile exists for given wallet user.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Boolean&gt;</code> - - true if profile is found.  
<a name="module_wallet-user--exports.WalletUser.WalletUser+unlock"></a>

#### exports.WalletUser.unlock(options) ⇒ <code>Promise.&lt;Object&gt;</code>
Unlocks wallet and returns a authorization token to be used for performing wallet operations.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - 'object.token' - auth token subsequent use of wallet features.  

| Param | Type | Description |
| --- | --- | --- |
| options | <code>Object</code> |  |
| options.localKMSPassphrase | <code>String</code> | (optional) passphrase for local kms for key operations. |
| options.webKMSAuth | <code>Object</code> | (optional) WebKMSAuth for authorizing access to web/remote kms. |
| options.webKMSAuth.authToken | <code>String</code> | (optional) Http header 'authorization' bearer token to be used. |
| options.webKMSAuth.capability | <code>String</code> | (optional) Capability if ZCAP sign header feature to be used for authorizing access. |
| options.webKMSAuth.authzKeyStoreURL | <code>String</code> | (optional) authz key store URL if ZCAP sign header feature to be used for authorizing access. |
| options.webKMSAuth.secretShare | <code>String</code> | (optional) secret share if ZCAP sign header feature to be used for authorizing access. |
| options.edvUnlocks | <code>Object</code> | (optional) for authorizing access to wallet's EDV content store. |
| options.edvUnlocks.authToken | <code>String</code> | (optional) Http header 'authorization' bearer token to be used. |
| options.edvUnlocks.capability | <code>String</code> | (optional) Capability if ZCAP sign header feature to be used for authorizing access. |
| options.edvUnlocks.authzKeyStoreURL | <code>String</code> | (optional) authz key store URL if ZCAP sign header feature to be used for authorizing access. |
| options.edvUnlocks.secretShare | <code>String</code> | (optional) secret share if ZCAP sign header feature to be used for authorizing access. |
| options.expiry | <code>Time</code> | (optional) time duration in milliseconds for which this profile will be unlocked. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+lock"></a>

#### exports.WalletUser.lock() ⇒ <code>Promise.&lt;Bool&gt;</code>
locks wallet by invalidating previously issued wallet auth.
Wallet has to be unlocked again to perform any future wallet operations.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Bool&gt;</code> - -  bool flag false if token is not found or already expired for this wallet user.  
<a name="module_wallet-user--exports.WalletUser.WalletUser+savePreferences"></a>

#### exports.WalletUser.savePreferences(auth, preferences) ⇒ <code>Promise.&lt;Object&gt;</code>
Saves TrustBloc wallet user preferences.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| preferences | <code>Object</code> |  |
| preferences.name | <code>String</code> | (optional) wallet user display name. |
| preferences.description | <code>Object</code> | (optional) wallet user display description. |
| preferences.image | <code>String</code> | (optional)  wallet user display image in URL format. |
| preferences.controller | <code>String</code> | (optional) default controller to be used for digital proof for this wallet user. |
| preferences.controllerPublished | <code>Boolean</code> | (optional) represents whether controller is published or not. |
| preferences.verificationMethod | <code>Object</code> | (optional) default verificationMethod to be used for digital proof for this wallet user. |
| preferences.proofType | <code>String</code> | (optional) default proofType to be used for digital proof for this wallet user. |
| preferences.proofFormat | <code>String</code> | (optional) default format (ldp, jwt) to be used for digital proof for this wallet user. |
| preferences.skipWelcomeMsg | <code>Boolean</code> | (optional) represents whether this wallet user has dismissed a welcome message in the UI. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+updatePreferences"></a>

#### exports.WalletUser.updatePreferences(auth, preferences) ⇒ <code>Promise.&lt;Object&gt;</code>
Updates TrustBloc wallet user preferences.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| preferences | <code>Object</code> |  |
| preferences.name | <code>String</code> | (optional) wallet user display name. |
| preferences.description | <code>Object</code> | (optional) wallet user display description. |
| preferences.image | <code>String</code> | (optional)  wallet user display image in URL format. |
| preferences.controller | <code>String</code> | (optional) default controller to be used for digital proof for this wallet user. |
| preferences.controllerPublished | <code>Boolean</code> | (optional) represents whether controller is published or not. |
| preferences.verificationMethod | <code>Object</code> | (optional) default verificationMethod to be used for digital proof for this wallet user. |
| preferences.proofType | <code>String</code> | (optional) default proofType to be used for digital proof for this wallet user. |
| preferences.proofFormat | <code>String</code> | (optional) default format (ldp, jwt) to be used for digital proof for this wallet user. |
| preferences.skipWelcomeMsg | <code>Boolean</code> | (optional) represents whether this wallet user has dismissed a welcome message in the UI. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+getPreferences"></a>

#### exports.WalletUser.getPreferences(auth) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets TrustBloc wallet user preference.

If controller not published, then this function checks if that controller is published.
If published then it change published to true in underlying wallet content store and updates user preference.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - promise containing preference metadata or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+saveMetadata"></a>

#### exports.WalletUser.saveMetadata(auth, content) ⇒ <code>Promise.&lt;Object&gt;</code>
Saves custom metadata data model into wallet.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - - empty promise or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| content | <code>Object</code> | metadata to be saved in wallet content store. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+getMetadata"></a>

#### exports.WalletUser.getMetadata(auth, contentID) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets metadata by ID from wallet.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - result.content - promise containing metadata or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |
| contentID | <code>String</code> | ID of the metadata to be read from wallet content store. |

<a name="module_wallet-user--exports.WalletUser.WalletUser+getAllMetadata"></a>

#### exports.WalletUser.getAllMetadata(auth) ⇒ <code>Promise.&lt;Object&gt;</code>
Gets All metadata data models from wallet.

**Kind**: instance method of [<code>exports.WalletUser</code>](#exp_module_wallet-user--exports.WalletUser)  
**Returns**: <code>Promise.&lt;Object&gt;</code> - result.contents - promise containing result or error if operation fails.  

| Param | Type | Description |
| --- | --- | --- |
| auth | <code>String</code> | authorization token for wallet operations. |

<a name="normalizePresentationSubmission"></a>

## normalizePresentationSubmission ⇒ <code>Array.&lt;Object&gt;</code>
Scans through @see [Presentation Submission](https://identity.foundation/presentation-exchange/#presentation-submission) descriptor map and groups results by descriptor IDs [descriptor_id -> Array of verifiable Credentials].
 In many cases, a single input descriptor can find multiple credentials.
 So this function is useful in cases of grouping results by input descriptor ID and giving option to user to choose only one to avoid over sharing.

 TODO: support for path_nested in descriptor map.

**Kind**: global constant  
**Returns**: <code>Array.&lt;Object&gt;</code> - - Normalized result array with each objects containing input descriptor id, name, purpose, format and array of credentials.  

| Param | Type | Description |
| --- | --- | --- |
| query | <code>Array.&lt;Object&gt;</code> | array of query, one of which could have produced the presentation. |
| presentation | <code>Object</code> | presentation response of query. If `presentation_submission` is missing, then normalization will treat this as non presentation exchange query  and normalization logic will only flatten the credentials (grouping duplicate results logic won't be applied). |

<a name="updatePresentationSubmission"></a>

## updatePresentationSubmission
Updates given presentation submission presentation by removing duplicate descriptor map entries.

 Descriptor map might contain single input descriptor ID mapped to multiple credentials.
 So returning PresentationSubmission presentation will retain only mappings mentioned in updates Object{<inputDescriptorID>:<credentialID>} parameter.

**Kind**: global constant  
<a name="findAttachmentByFormat"></a>

## findAttachmentByFormat
Finds attachment by given format.
 Supporting Attachment Format from DIDComm V1.

 Note: Currently finding only one attachment per format.

**Kind**: global constant  
<a name="findAttachmentByFormatV2"></a>

## findAttachmentByFormatV2
Finds attachment by given format.
 Supporting Attachment Format from DIDComm V2.

 Note: Currently finding only one attachment per format.

**Kind**: global constant  
<a name="extractOOBGoalCode"></a>

## extractOOBGoalCode
Reads out-of-band invitation goal code.
 Supports DIDComm V1 & V2

**Kind**: global constant  
<a name="waitFor"></a>

## waitFor()
Wait for given duration in millisecond and return promise.

**Kind**: global function  
## Contributing
Thank you for your interest in contributing. Please see our [community contribution guidelines](https://github.com/trustbloc/community/blob/main/CONTRIBUTING.md) for more information.

## License
Apache License, Version 2.0 (Apache-2.0). See the [LICENSE](LICENSE) file.
