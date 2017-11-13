package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
)

type PluginConf struct {
	types.NetConf
	RuntimeConfig *struct {
		MasqEntries []MasqEntry `json:"masqEntries,omitempty"`
	} `json:"runtimeConfig,omitempty"`

	RawPrevResult *map[string]interface{} `json:"prevResult"`
	PrevResult    *current.Result         `json:"-"`

	Debug    bool   `json:"debug"`
	DebugDir string `json:"debugDir"`
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

	if conf.Debug == true && conf.DebugDir == "" {
		return nil, fmt.Errorf("debugDir must be specified")
	}

	for _, e := range conf.RuntimeConfig.MasqEntries {
		if !e.Valid() {
			return nil, fmt.Errorf("Invalid Masq entry")
		}
	}

	return &conf, nil
}

func cniDebug(enabled bool, dir string, args *skel.CmdArgs, action string) {
	if !enabled {
		return
	}
	dFilePath := filepath.Join(dir, args.ContainerID, args.IfName)
	if err := os.MkdirAll(dFilePath, 0770); err != nil {
		panic("untested")
	}
	stdinFile := fmt.Sprintf("%s/%s_%v.json", dFilePath, action, time.Now().Unix())
	if err := ioutil.WriteFile(stdinFile, args.StdinData, 0770); err != nil {
		panic("untested")
	}

	netnsFile := fmt.Sprintf("%s/%s_%v.netns", dFilePath, action, time.Now().Unix())
	if err := ioutil.WriteFile(netnsFile, []byte(args.Netns), 0770); err != nil {
		panic("untested")
	}
}

func cmdAdd(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	cniDebug(conf.Debug, conf.DebugDir, args, "add")

	//If first do nothing?
	if conf.PrevResult == nil {
		return types.PrintResult(&current.Result{}, conf.CNIVersion)
	}

	for _, e := range conf.RuntimeConfig.MasqEntries {
		ip, err := getContainerIP(conf.PrevResult, args)
		if err != nil {
			panic("untestedx")
		}
		if err := e.Apply(ip, "mymasq", "mycomment"); err != nil {
			panic(err.Error())
		}
	}

	return types.PrintResult(conf.PrevResult, conf.CNIVersion)
}

func applyMasqRules(rules, ip string) error {
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

func cmdDel(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}
	_ = conf

	cniDebug(conf.Debug, conf.DebugDir, args, "del")

	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports(version.Current()))
}
