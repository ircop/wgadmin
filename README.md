# wgadmin: configuration sync/modify tool for distributed wireguard server installations

Motivation: most of VPN servers have ability to authorize clients via radius: either builtin support or via plugins. But wireguard doesn't.
Nevertheless, WG is great vpn protocol and many people love to use it.

This tool allows wireguard server owners, which have multiple wg servers and want to change wg configs on the fly, to administrate their 
wg servers via simple http(s) api (or embed wgadmin/master or wgadmin/slave packages in their own golang software)


# Master daemon

```shell script
go install github.com/ircop/wgadmin/cmd/wgmaster
```

Master daemon starts https server on given addr/port, and accepts slaves connections on '/wg' endpoint, as well as API calls.

Only simplest basic authorization allowed by setting LOGIN/PASSWORD env vars. Use reverse proxy for complex auth.




```shell script
LOGIN=login PASSWORD=pw HOST=0.0.0.0 PORT=4321 \
    CERT=/path/to/ssl.crt KEY=/path/to/ssl.key \
    $GOPATH/bin/wgmaster
```
Possible env vars:

LOGIN, PASSWORD: basic auth credentials. Basic auth disabled, if login or password not set. If you need more complex auth, use nginx with htpasswd as reverse proxy.

HOST, PORT: host/port to listen on

CERT, KEY: ssl cert/key.

# Slave daemon

Slave daemon runs on each wireguard node, connects to master with given login/password, if needed (basic auth),
and accepts commands for add/remove/sync wireguard peers. 

```shell script
go install github.com/ircop/wgadmin/cmd/wgslave

SAVE_TEMPLATE=/etc/wireguard/template.conf SAVE_PATH=/etc/wireguard/clients.conf \
  LOGIN=t1 PASSWORD=t2 \
  REMOTE=127.0.0.1:4321/wg \
  IFNAME=clients SKIPTLSVERIFY=true \
  $GOPATH/bin/wgslave
```

Configuration saving:

By default, configuration changes are in-memory only. If SAVE_TEMPLATE and SAVE_PATH are passed, config will be saved
at SAVE_PATH after each change.

Possible env vars:

SAVE_TEMPLATE: template header for wireguard configuration. Usually there is [Interface] config section.

SAVE_PATH: path to configuration file. On each change, SAVE_TEMPLATE will be merged with in-memory [Peers] and saved to given path.

LOGIN, PASSWORD: login and password for basic auth (if needed)

REMOTE: host:port/endpoint where to connect. At now endpoint is always `/wg`, and host/port are configurable at master's launch

IFNAME: wireguard interface name

SKIPTLSVERIFY: `warning: insecure option`. Skip TLS verify during master connection. Useful for dealing with self-signed certificates, for testing purposes, etc.



# api usage overview

List of connected peers (slaves):
```shell script
$ curl -k https://localhost:4321/peers  | jq
[
  "10.10.10.1",
  "10.10.10.2"
]
```

List of slave wireguard peers configured:
```shell script
$ curl -k https://localhost:4321/peer/10.10.10.1  | jq
{
  "peers": [
    {
      "key": "DX7OxfYr72A0SVBt7IwAgfA8xGpdYqz5fAbhiRS5YSI=",
      "ip": "10.20.0.2"
    },
    {
      "key": "0ImwTBsnV54hh/KnutTJeCcKxQh1L5qd52Y8UEjNbGk=",
      "ip": "10.20.0.3"
    }
  ]
}
```

Add peer to slave's config:
```shell script
$ curl -k -X POST https://localhost:4321/peer/10.10.10.1 -d '[{"ip":"10.10.10.200","key":"MFmKiO60hlCfXr+Sd0pRAC8+gjlbyoM+UQWeHvHRomY="}]'
"OK"
```

Add peer to all slaves:
```shell script
$ curl -k -X POST https://localhost:4321/all -d '[{"ip":"10.10.10.100","key":"mBwYcYy4XED9lfNHgac6NWkjB1YhGV8sCAxHZo5bo0k="}]' | jq
{
  "results": {
    "10.10.10.1": "ok",
    "10.10.10.2": "some error message"
  }
}
```

Remove peer from slave:
```shell script
$ curl -k -X POST https://localhost:4321/peer/10.10.10.1/delete -d '["DX7OxfYr72A0SVBt7IwAgfA8xGpdYqz5fAbhiRS5YSI="]'
"OK"
```

Remove peer from all slaves:
```shell script
$ curl -k -X POST https://localhost:4321/all/delete -d '["MFmKiO60hlCfXr+Sd0pRAC8+gjlbyoM+UQWeHvHRomY="]' | jq
{
  "results": {
    "10.10.10.1": "ok",
    "10.10.10.2": "ok"
  }
}
```

Sync peer config: replace current peers with new ones.
```shell script
$ curl -k -X POST https://localhost:4321/peer/10.10.10.1/sync -d \
  '[{"ip":"1.2.3.8","key":"CFKqxwJlotKtKLSk/S85IDdKJtWINZjOTK1WcrJZek0="},\
  {"ip":"7.6.5.4","key":"gPX2BQ9K1MB3ZBuejWXUjW37ea4E5/Hj6kyJLfPYXEo="}]'
"OK"
```

Sync all peers:
```shell script
$ curl -k -X POST https://localhost:4321/all/sync -d \
  '[{"ip":"1.2.3.8","key":"CFKqxwJlotKtKLSk/S85IDdKJtWINZjOTK1WcrJZek0="},\
  {"ip":"7.6.5.4","key":"gPX2BQ9K1MB3ZBuejWXUjW37ea4E5/Hj6kyJLfPYXEo="}]'
{
  "results": {
    "10.10.10.1": "ok",
    "10.10.10.2": "ok"
  }
}
```


# Using as package

You can use this as a packages in your own software. Just import `github.com/ircop/wgadmin/wglib/slave` 
or `github.com/ircop/wgadmin/wglib/master` in your code.

See example usage in cmd/master and cmd/slave
