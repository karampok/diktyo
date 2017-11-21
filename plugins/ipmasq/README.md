### ipmasq plugin 

This plugin will modify the external traffic. 
It expects to be run as a chained plugin.

//iptables -t nat -A POSTROUTING -p tcp -o eth0 -j SNAT --to 1.2.3.4:1-1023

```
unset CNI_COMMAND CNI_IFNAME CNI_NETNS CNI_CONTAINERID
export CNI_PATH=/opt/cni/bin
export NETCONFPATH=/opt/cni/netconfs
mkdir -p {$NETCONFPATH,$CNI_PATH}
cp {cnitool,bridge,host-local,ipmasq} $CNI_PATH
ip netns add bob||true
```

## Usage
You should use this plugin as part of a network configuration list. It accepts
the following configuration options:

* `tag` - boolean, default true. If true or omitted,

The plugin expects to receive the actual list of port mappings via the 
`masqEntries` [capability argument](https://github.com/containernetworking/cni/blob/master/CONVENTIONS.md)


So a sample standalone config list (with the file extension .conflist) might
look like:
```json
cat > $NETCONFPATH/chained.conflist <<EOF
{
  "cniVersion": "0.3.1",
  "name": "mycoolnet",
  "plugins": [
     {
        "type": "bridge",
        "isGateway": true,
        "ipMasq": false,
        "bridge": "mycoolbridge",
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
      "tag":"mycoolChain",
      "capabilities": {"masqEntries": true,"metadata": true}
    }
  ]
}
EOF
```


Runtime engine is expecting to insert entries through the runtime config
```
export CAP_ARGS='{
    "masqEntries": [
        {
            "external":      "10.0.2.15:4000-4010",
            "protocol":      "tcp",
            "description":   "allow production traffic"
        }
    ]
}'

```


cnitool add mycoolnet /var/run/netns/bob
