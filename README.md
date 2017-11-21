# Diktyo CNI plugin

A collection of CNI plugins 

## Plugins

* ipmasq
* noop


## Demo

```
unset CNI_COMMAND CNI_IFNAME CNI_NETNS CNI_CONTAINERID
export CNI_PATH=/opt/cni/bin
export NETCONFPATH=/opt/cni/netconfs
mkdir -p {$NETCONFPATH,$CNI_PATH}
cp $GOBIN/{cnitool,bridge,portmap,host-local,ipmasq,noop} $CNI_PATH
ip netns add bob||true
```

So a sample standalone config list (with the file extension .conflist) might
look like:

```json
cat > $NETCONFPATH/chained.conflist <<EOF
{
  "cniVersion": "0.3.1",
  "name": "mycoolcnichain",
  "plugins": [
     {
        "type": "bridge",
        "isGateway": true,
        "ipMasq": false,
        "bridge": "mybridge",
        "ipam": {
            "type": "host-local",
            "subnet": "10.10.30.0/24",
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
      "type":"ipmasq",
      "tag":"CNI-SNAT-X",
      "capabilities": {"masqEntries": true}
    },    
    {
      "type": "portmap",
      "capabilities": {"portMappings": true},
      "snat": false
    },
    {
      "type":"noop",
      "debug":true,
      "capabilities": {
           "portMappings": true,
           "masqEntries": true
      },
      "debugDir": "/tmp/net-debug"
    }
  ]
}
EOF
```

Container engine is expecting to insert entries through the runtime config
(cni-tool does that through CAP_ARGS)

```
export CAP_ARGS='{
    "portMappings": [
        {
          "hostPort": 60001,
          "protocol":      "tcp",
          "containerPort": 8080
        }, {
          "hostPort": 60002,
          "protocol":      "tcp",
          "containerPort": 2222
        }
    ],
    "masqEntries": [
        {
            "external":      "10.0.2.15:5000-5010",
            "destination":   "8.8.8.8/32",
            "protocol":      "tcp",
            "description":   "allow production traffic"
        },
        {
            "external":      "10.0.2.15:4000-4010",
            "destination":   "0.0.0.0/0",
            "protocol":      "tcp",
            "description":   "allow production traffic"
        }
    ]
}'
```

Finally
```
cnitool add mycoolcnichain /var/run/netns/bob
#cat /tmp/net-debug/cnitool-XYZ/eth0/add_date.json | jq .
iptables -nvL -t nat
cnitool add mycoolcnichain /var/run/netns/bob
```


