### iproute plugin

This plugin will modify the routing table within the container network namespace.
It expects to be run as a chained plugin.

```
unset CNI_COMMAND CNI_IFNAME CNI_NETNS CNI_CONTAINERID
export CNI_PATH=/opt/cni/bin
export NETCONFPATH=/opt/cni/netconfs
mkdir -p {$NETCONFPATH,$CNI_PATH}
cp $GOBIN/{cnitool,bridge,host-local,iproute} $CNI_PATH
ip netns add bob||true
```

## Usage

So a sample standalone config list (with the file extension .conflist) might
look like:

```json
cat > $NETCONFPATH/chained.conflist <<EOF
{
  "cniVersion": "0.3.1",
  "name": "mynet",
  "plugins": [
     {
        "type": "bridge",
        "isGateway": true,
        "ipMasq": false,
        "bridge": "mybridge",
        "ipam": {
            "type": "host-local",
            "subnet": "10.10.0.0/16",
            "routes": [
                { "dst": "0.0.0.0/0" }
            ],
         "dataDir": "/run/ipam-out-net"
        },
        "dns": {
          "nameservers": [ "8.8.8.8" ]
        }
    },
    {
      "type":"iproute",
      "capabilities": {"routeEntries": true}
    }
  ]
}
EOF
```

Runtime engine is expecting to insert entries through the runtime config

```json
export CAP_ARGS='{
    "routeEntries": [
        {
            "destination":   "8.8.8.8/32",
            "gateway":      "drop",
            "description":   "do not access google dns server"
        },
        {
            "destination":   "192.168.0.0/16",
            "gateway":      "10.10.30.1",
            "description":   "VPN traffic"
        }
    ]
}'
```

```
cnitool add mynet /var/run/netns/bob

# $ ip netns exec bob ip r
# default via 10.10.30.1 dev eth0
# blackhole 8.8.0.0/16
# blackhole 8.8.8.8
# 10.10.30.0/24 dev eth0  proto kernel  scope link  src 10.10.30.13

```
