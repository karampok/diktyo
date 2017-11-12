package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	if conf.PrevResult == nil {
		conf.PrevResult = &current.Result{}
	}

	return types.PrintResult(conf.PrevResult, conf.CNIVersion)
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
