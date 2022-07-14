# Agent JS Worker

Agent SDK via Javascript.

## Build it

To build you will need:

* Go 1.17.x
* npm 6.13.x
* Node.JS 12.14.x
* bash

Run `npm install && npm run build` in this directory. The output bundles will be placed in `dist/`.

## Usage

> Note: the API is in the very early stages of development and is still subject to a few changes.

### Entrypoints

`agent-js-worker` has several entrypoints tailored to the environment and needs:

* `dist/web/agent.js`: for use in the browser
* `dist/rest/agent.js`: for use in any environment but relying on an external
  [REST controller API server](https://github.com/trustbloc/agent-sdk/blob/master/docs/rest/README.md)
  instead of the bundled webassembly module.

### Snippet

**Example:** accept a did-exchange invitation:

```js
// in the browser (this agent's initialization shows all possible arguments)

const agent = await new Agent.Framework({
    assetsPath: "/public/dist/assets",
    "agent-default-label": "dem-js-agent",
    "http-resolver-url": [],
    "auto-accept": true,
    "outbound-transport": ["ws", "http"],
    "transport-return-route": "all",
    "log-level": "debug",
    "storageType": "indexedDB",
    "indexedDB-namespace": "demo",
    "edvServerURL": "",
    "edvVaultID": "",
    "edvCapability": "",
    "blocDomain": "",
    "trustbloc-resolver": "",
    "authzKeyStoreURL": "",
    "opsKeyStoreURL": "",
    "edvOpsKIDURL": "",
    "edvHMACKIDURL": "",
    "kmsType": "",
    "userConfig": {
        "walletSecretShare": ""
    },
    "useEDVCache": false,
    "edvClearCache": "",
    "useEDVBatch": false,
    "edvBatchSize": 0,
    "cacheSize": 0
})

// sample invitation
const invitation = {
    "@id":"4d26ad47-c71b-4e2e-9358-0a76f7fa77e4",
    "@type":"https://didcomm.org/didexchange/1.0/invitation",
    "label":"demo-js-agent",
    "recipientKeys":["7rADm5sA9FHB4enuYXj6PJZDAm1JcesKmbtx7Qh8YZrg"],
    "serviceEndpoint":"routing:endpoint"
};

// listen for connection 'received' notification
agent.startNotifier(notice => {
    const event = notice.payload
    if (event.Type === "post_state") {
        // accept invitation
        agent.didexchange.acceptInvitation(event.Properties.connectionID)
    }
}, ["didexchange_states"])
// receive invitation
agent.didexchange.receiveInvitation(invitation)

// listen for connection 'completed' notification
agent.startNotifier(notice => {
    const event = notice.payload
    if (event.StateID === "completed" && event.Type === "post_state") {
        console.log("connection completed!")
    }

}, ["didexchange_states"])

// release resources
agent.destroy()
```

### Browser

Note: make sure the assets are [served correctly](#important---serving-the-assets).

Source `agent.js` in your `<script>` tag:

```html
<script src="dist/web/agent.js"></script>
```

Then initialize your agent instance:

```js
const agent = await new Agent.Framework({
    assetsPath: "/public/dist/assets",
    "agent-default-label": "dem-js-agent",
    "http-resolver-url": [],
    "auto-accept": true,
    "outbound-transport": ["ws", "http"],
    "transport-return-route": "all",
    "indexedDB-namespace": "demo",
    "storageType": "indexedDB",
    "log-level": "debug"
})
```

### REST

Note: make sure the assets are [served correctly](#important---serving-the-assets) if you're running agent in the browser.

Assuming you're in the browser, source `agent.js` in your `<script>` tag:

```html
<script src="dist/rest/agent.js"></script>
```

Then initialize your agent instance:

```js
const agent = await new Agent.Framework({
    assetsPath: "/path/serving/the/assets", // still required for assets other than the wasm
    "agent-rest-url": "http://controller.api.example.com", // REST controller URL of the agent
    "agent-rest-wshook": "ws://controller.api.example.com", // Optional REST controller websocket URL from which you can listen to notifications
    "agent-rest-token": "sample_auth_token" // Optional authorization header to be based to rest endpoint for each request
})
```

### Important - Serving the Assets

Note: this applies if you are running in the browser.

`agent-js-worker` loads some assets at runtime: the web assembly binary and a couple of JS scripts. These assets are
located in the `dist/assets` directory (if you `npm install` it, you'll find them in
`./node_modules/@trustbloc/agent-sdk-web/dist/assets`).

Things that need to work if you are to use `agent-js-worker` on the client side:

#### Headers

Make sure the content server adds the appropriate headers when serving the compressed `agent-lite-js-worker.wasm` file.
`agent-js-worker` uses the [Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) to fetch the wasm.

Examples:

**Serving gzipped wasm:**

Headers:

* `Content-Type: application/wasm`
* `Content-Encoding: gzip`

**Serving wasm compressed with brotli:**

If your browser supports it, then the headers are:

* `Content-Type: application/wasm`
* `Content-Encoding: br`

Note, however, that your browser may not support this compression mode.
 
Not all browsers include `br` in `Accept-Encoding` when using `fetch()` (Firefox doesn't) and it is impossible to
override because `Accept-Encoding` is a [forbidden header name](https://fetch.spec.whatwg.org/#forbidden-header-name).

**Serving uncompressed wasm (not recommended):**

Headers:

* `Content-Type: application/wasm`

#### Path

The URL used to fetch the WASM file is **always** `<assetsPath>/agent-lite-js-worker.wasm`.
This path needs to exist even if your content server is serving a compressed version.

#### Configuring your content server

Here are some examples:

**Nginx**

[Sending compressed files](https://docs.nginx.com/nginx/admin-guide/web-server/compression/#sending-compressed-files):
enabling `gzip_static` on a location will automatically serve requests to `http://example.com/assets/agent-lite-js-worker.wasm`
with `agent-lite-js-worker.wasm.gz` if it exists.

Example: Nginx serving your assets under `/public/assets` with gzipped wasm:

```
location ~ agent-js-worker\.wasm$ {
    gzip_static on;
    types {
        application/wasm  wasm;
    }
}
```

Files in `/public/assets`:

```
assets
├── agent-lite-js-worker.wasm.gz
├── wasm_exec.js
├── worker-impl-rest.js
└── worker-impl-web.js
```

Requests for `http://example.com/public/assets/agent-lite-js-worker.wasm` will be served with the `.gz` file.

**goexec**

Here is a hacky one-liner when using [`goexec`](https://github.com/shurcooL/goexec) (for development purposes):

```
goexec -quiet 'http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {dir := http.Dir("."); if strings.HasSuffix(r.RequestURI, ".wasm") && !strings.Contains(r.RequestURI, "wasm=") {w.Header().Add("Content-Encoding", "gzip"); w.Header().Add("Content-Type", "application/wasm"); fmt.Sprintf(r.URL.Path); file, err := dir.Open(r.URL.Path + ".gz"); if err != nil {w.Header().Add("x-error", err.Error()); w.WriteHeader(http.StatusInternalServerError); return; }; buf := make([]byte, 2048); for err == nil { n := 0; n, err = file.Read(buf);if n > 0 {n, err = w.Write(buf[:n]);}}; if !errors.Is(err, io.EOF) {w.WriteHeader(http.StatusInternalServerError); return;}; }; http.FileServer(http.Dir(".")).ServeHTTP(w, r) }); http.ListenAndServe(":8080", nil)'
```
