## registerDNS plugin 

This plugin will register a DNS entry to a plugable. 
It expects to be run as a chained plugin.
It expects to be a binary that is responsible to register the DNS entry.



## Usage
You should use this plugin as part of a network configuration list. It accepts
the following configuration options:

* `plugin` - string, which binary to use
* `app_id_key` - string, which field in the metadata to be used //TODO. make more that one

Optionally, the plugin expects to receive the actual list of metadata along with the container handle


## How it works with consul

Install & Run Consul (  [link](https://gist.github.com/karampok/171701458b14c394387d359429197695) 
Install & Check the DNS binary plugin

```
wget https://github.com/mantl/consul-cli/releases/download/v0.3.1/consul-cli_0.3.1_linux_amd64.tar.gz
tar xvf consul-cli_0.3.1_linux_amd64.tar.gz
mv consul-cli_0.3.1_linux_amd64/consul-cli /usr/local/bin/
sudo mv consul-cli_0.3.1_linux_amd64/consul-cli /usr/local/bin/
consul-cli agent members
consul-cli service  register cake --address="1.2.3.4"
dig @127.0.0.1 -p 8600 cake.service.consul
```

Create a network namespace and the configs

```
unset CNI_COMMAND CNI_IFNAME CNI_NETNS CNI_CONTAINERID
export CNI_PATH=/opt/cni/bin
export NETCONFPATH=/opt/cni/netconfs
mkdir -p {$NETCONFPATH,$CNI_PATH}
cp $GOBIN/{cnitool,bridge,host-local,registerDNS} $CNI_PATH
ip netns add bob||true
```

So a sample standalone config list looks like 
look like:

```
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
      "type":"registerDNS",
      "plugin":"consul-cli",
      "app_id_key":"app_id",
      "capabilities": {"metadata": true}
    }
  ]
}
EOF
```

Container engine is expecting to insert entries through the runtime config

```
export CAP_ARGS='{
    "metadata": 
        {
            "spaceID":    "spaceY",
            "app_id":      "appX"
        }
}'
```

Create two instances of containers

```
export CNI_CONTAINERID=cake1 CNI_IFNAME=eth0
cnitool add mycoolnet /var/run/netns/bob
export CNI_CONTAINERID=cake2 CNI_IFNAME=eth1
cnitool add mycoolnet /var/run/netns/bob
dig @127.0.0.1 -p 8600 appX.service.consul
```


