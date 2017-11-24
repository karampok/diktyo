package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

type PluginConf struct {
	types.NetConf
	RuntimeConfig *struct {
		Metadata *map[string]interface{} `json:"metadata"`
	} `json:"runtimeConfig,omitempty"`

	RawPrevResult *map[string]interface{} `json:"prevResult"`
	PrevResult    *current.Result         `json:"-"`

	Plugin   string `json:"plugin"`
	AppIDKey string `json:"app_id_key"`
}

func parseConfig(stdin []byte) (*PluginConf, error) {
	conf := PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	if conf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(conf.RawPrevResult)
		if err != nil {
			return nil, fmt.Errorf("could not serialize prevResult: %v", err)
		}
		res, err := version.NewResult(conf.CNIVersion, resultBytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse prevResult: %v", err)
		}
		conf.RawPrevResult = nil
		conf.PrevResult, err = current.NewResultFromResult(res)
		if err != nil {
			return nil, fmt.Errorf("could not convert result to current version: %v", err)
		}
	}

	return &conf, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	if conf.PrevResult == nil {
		return fmt.Errorf("must be called as chained plugin")
	}

	ip, err := getContainerIP(conf.PrevResult, args)
	if err != nil {
		panic("untestedx")
	}

	if conf.RuntimeConfig != nil && conf.RuntimeConfig.Metadata != nil {
		resultBytes, err := json.Marshal(conf.RuntimeConfig.Metadata)
		if err != nil {
			return fmt.Errorf("could not serialize prevResult: %v", err)
		}
		var meta map[string]interface{}
		if err := json.Unmarshal(resultBytes, &meta); err != nil {
			return fmt.Errorf("could not parse metadata: %v", err)
		}
		if v, ok := meta[conf.AppIDKey]; ok {
			cmd := exec.Command(conf.Plugin, "service", "register", v.(string),
				fmt.Sprintf("--id=%s", args.ContainerID), fmt.Sprintf("--address=%s", ip),
			)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("calling the plugin to register: %s", err)
			}
		}
	}

	return types.PrintResult(conf.PrevResult, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	cmd := exec.Command(conf.Plugin, "service", "deregister", args.ContainerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("calling the plugin to register: %s", err)
	}

	return nil
}

func getContainerIP(c *current.Result, args *skel.CmdArgs) (net.IP, error) {
	containerIPs := make([]net.IP, 0, len(c.IPs))
	if c.CNIVersion != "0.3.0" {
		for _, ip := range c.IPs {
			containerIPs = append(containerIPs, ip.Address.IP)
		}
	} else {
		for _, ip := range c.IPs {
			if ip.Interface == nil {
				continue
			}
			intIdx := *ip.Interface
			// Every IP is indexed in to the interfaces array, with "-1" standing
			// for an unknown interface (which we'll assume to be Container-side
			// Skip all IPs we know belong to an interface with the wrong name.
			if intIdx >= 0 && intIdx < len(c.Interfaces) && c.Interfaces[intIdx].Name != args.IfName {
				continue
			}
			containerIPs = append(containerIPs, ip.Address.IP)
		}
	}
	if len(containerIPs) == 0 {
		return nil, fmt.Errorf("got no container IPs")
	}

	return containerIPs[0], nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports(version.Current()))
}
