# Run the agent as a binary

## Build the Agent

The agent can be built from within the `cmd/agent-rest` directory with `go build`.

## Run the Agent

Start the agent with `./agent-rest start [flags]`.

## Agent Parameters

Parameters can be set by command line arguments or environment variables:

```
Flags:
  -l, --agent-default-label string         Default Label for this agent. Defaults to blank if not set. Alternatively, this can be set with the following environment variable: ARIESD_DEFAULT_LABEL
  -a, --api-host string                    Host Name:Port. Alternatively, this can be set with the following environment variable: ARIESD_API_HOST
  -t, --api-token string                   Check for bearer token in the authorization header (optional). Alternatively, this can be set with the following environment variable: ARIESD_API_TOKEN
      --auto-accept string                 Auto accept requests. Possible values [true] [false]. Defaults to false if not set. Alternatively, this can be set with the following environment variable: ARIESD_AUTO_ACCEPT
  -u, --database-prefix string             An optional prefix to be used when creating and retrieving underlying databases.  Alternatively, this can be set with the following environment variable: ARIESD_DATABASE_PREFIX
      --database-timeout string            Total time in seconds to wait until the db is available before giving up. Default: 30 seconds. Alternatively, this can be set with the following environment variable: ARIESD_DATABASE_TIMEOUT
  -q, --database-type string               The type of database to use for everything except key storage. Supported options: mem, couchdb, mysql, leveldb, mongodb.  Alternatively, this can be set with the following environment variable: ARIESD_DATABASE_TYPE
  -v, --database-url string                The URL (or connection string) of the database. Not needed if using memstore. For CouchDB, include the username:password@ text if required.  Alternatively, this can be set with the following environment variable: ARIESD_DATABASE_URL
  -h, --help                               help for start
  -r, --http-resolver-url method@url       HTTP binding DID resolver method and url. Values should be in method@url format. This flag can be repeated, allowing multiple http resolvers. Defaults to peer DID resolver if not set. Alternatively, this can be set with the following environment variable (in CSV format): ARIESD_HTTP_RESOLVER
  -i, --inbound-host scheme@url            Inbound Host Name:Port. This is used internally to start the inbound server. Values should be in scheme@url format. This flag can be repeated, allowing to configure multiple inbound transports. Alternatively, this can be set with the following environment variable: ARIESD_INBOUND_HOST
  -e, --inbound-host-external scheme@url   Inbound Host External Name:Port and values should be in scheme@url format This is the URL for the inbound server as seen externally. If not provided, then the internal inbound host will be used here. This flag can be repeated, allowing to configure multiple inbound transports. Alternatively, this can be set with the following environment variable: ARIESD_INBOUND_HOST_EXTERNAL
      --log-level string                   Log level. Possible values [INFO] [DEBUG] [ERROR] [WARNING] [CRITICAL] . Defaults to INFO if not set. Alternatively, this can be set with the following environment variable: ARIESD_LOG_LEVEL
  -o, --outbound-transport strings         Outbound transport type. This flag can be repeated, allowing for multiple transports. Possible values [http] [ws]. Defaults to http if not set. Alternatively, this can be set with the following environment variable: ARIESD_OUTBOUND_TRANSPORT
      --edv-server-url string              EDV server URL. Alternatively, this can be set with the following environment variable (in CSV format): ARIESD_EDV_SERVER_URL
  -c, --tls-cert-file string               tls certificate file. Alternatively, this can be set with the following environment variable: TLS_CERT_FILE
  -k, --tls-key-file string                tls key file. Alternatively, this can be set with the following environment variable: TLS_KEY_FILE
      --transport-return-route string      Transport Return Route option. Refer https://github.com/hyperledger/aries-framework-go/blob/8449c727c7c44f47ed7c9f10f35f0cd051dcb4e9/pkg/framework/aries/framework.go#L165-L168. Alternatively, this can be set with the following environment variable: ARIESD_TRANSPORT_RETURN_ROUTE
  -d, --trustbloc-domain string            Trustbloc domain URL. Alternatively, this can be set with the following environment variable (in CSV format): ARIESD_TRUSTBLOC_DOMAIN
      --trustbloc-resolver string          Trustbloc resolver URL. Alternatively, this can be set with the following environment variable (in CSV format): ARIESD_TRUSTBLOC_RESOLVER
  -w, --webhook-url strings                URL to send notifications to. This flag can be repeated, allowing for multiple listeners. Alternatively, this can be set with the following environment variable (in CSV format): ARIESD_WEBHOOK_URL
```

## Example

```shell
$ cd cmd/agent-rest
$ go build
$ ./agent-rest start start --database-type=mem --api-host localhost:8080 --inbound-host http@localhost:8081,ws@localhost:8082 --inbound-host-external http@https://example.com:8081,ws@ws://localhost:8082 --webhook-url localhost:8082 --agent-default-label MyAgent
```