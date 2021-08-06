---
title: "Configuration"
date: "2021-08-06T10:55:04+0000"
weight: 20
geekdocRepo: https://github.com/owncloud/ocis-wopiserver
geekdocEditPath: edit/main/templates
geekdocFilePath: CONFIGURATION.tmpl
---


{{< toc >}}

## Configuration

### Configuration using config files

Out of the box extensions will attempt to read configuration details from:

```console
/etc/ocis
$HOME/.ocis
./config
```

For this configuration to be picked up, have a look at your extension `root` command and look for which default config name it has assigned. *i.e: ocis-accounts reads `accounts.json | yaml | toml ...`*.

So far we support the file formats `JSON` and `YAML`, if you want to get a full example configuration just take a look at [our repository](https://github.com/owncloud/ocis/tree/master/accounts/config), there you can always see the latest configuration format. These example configurations include all available options and the default values. The configuration file will be automatically loaded if it's placed at `/etc/ocis/ocis.yml`, `${HOME}/.ocis/ocis.yml` or `$(pwd)/config/ocis.yml`.

### Environment variables

If you prefer to configure the service with environment variables you can see the available variables below.

If multiple variables are listed for one option, they are in order of precedence. This means the leftmost variable will always win if given.

### Commandline flags

If you prefer to configure the service with commandline flags you can see the available variables below. Command line flags are only working when calling the subcommand directly.

### accounts server

start wopiserver

Usage: `accounts server [command options] [arguments...]`







-log-file |  $WOPISERVER_LOG_FILE , $OCIS_LOG_FILE
: Enable log to file.


-tracing-enabled |  $WOPISERVER_TRACING_ENABLED
: Enable sending traces.


-tracing-type |  $WOPISERVER_TRACING_TYPE
: Tracing backend type. Default: `jaeger`.


-tracing-endpoint |  $WOPISERVER_TRACING_ENDPOINT
: Endpoint for the agent.


-tracing-collector |  $WOPISERVER_TRACING_COLLECTOR
: Endpoint for the collector.


-tracing-service |  $WOPISERVER_TRACING_SERVICE
: Service name for tracing. Default: `wopiserver`.


-debug-addr |  $WOPISERVER_DEBUG_ADDR
: Address to bind debug server. Default: `0.0.0.0:9109`.


-debug-token |  $WOPISERVER_DEBUG_TOKEN
: Token to grant metrics access.


-debug-pprof |  $WOPISERVER_DEBUG_PPROF
: Enable pprof debugging.


-debug-zpages |  $WOPISERVER_DEBUG_ZPAGES
: Enable zpages debugging.


-http-namespace |  $WOPISERVER_HTTP_NAMESPACE
: Set the base namespace for the http namespace. Default: `com.owncloud.web`.


-http-addr |  $WOPISERVER_HTTP_ADDR
: Address to bind http server. Default: `0.0.0.0:9105`.


-http-root |  $WOPISERVER_HTTP_ROOT
: Root path of http server. Default: `/`.


-http-cache-ttl |  $WOPISERVER_CACHE_TTL
: Set the static assets caching duration in seconds. Default: `604800`.


-name |  $WOPISERVER_NAME
: service name. Default: `"wopiserver"`.


-asset-path |  $WOPISERVER_ASSET_PATH
: Path to custom assets.


-wopi-server-host |  $WOPISERVER_WOPI_SERVER_HOST
: Wopiserver Host. Default: `http://127.0.0.1:8880`.


-wopi-server-insecure |  $WOPISERVER_WOPI_SERVER_INSECURE
: Wopiserver insecure. Default: `false`.


-wopi-server-iop-secret |  $WOPISERVER_WOPI_SERVER_IOP_SECRET
: shared IOP secret for CS3 WOPI server.


-wopi-server-token-ttl |  $WOPISERVER_TOKEN_TTL
: TTL of issued tokens. Default: `(1 * time.Hour)`.


-jwt-secret |  $WOPISERVER_JWT_SECRET , $OCIS_JWT_SECRET
: Used to create JWT to talk to reva, should equal reva's jwt-secret. Default: `"Pive-Fumkiu4"`.


-reva-gateway-addr |  $WOPISERVER_REVA_GATEWAY_ADDR
: Reva gateway address. Default: `"127.0.0.1:9142"`.

### accounts health

Check health status

Usage: `accounts health [command options] [arguments...]`






-debug-addr |  $WOPISERVER_DEBUG_ADDR
: Address to debug endpoint. Default: `0.0.0.0:9109`.























### accounts wopiserver

wopiserver, an example oCIS extension

Usage: `accounts wopiserver [command options] [arguments...]`


-config-file |  $WOPISERVER_CONFIG_FILE
: Path to config file.


-log-level |  $WOPISERVER_LOG_LEVEL , $OCIS_LOG_LEVEL
: Set logging level. Default: `info`.


-log-pretty |  $WOPISERVER_LOG_PRETTY , $OCIS_LOG_PRETTY
: Enable pretty logging.


-log-color |  $WOPISERVER_LOG_COLOR , $OCIS_LOG_COLOR
: Enable colored logging.
























