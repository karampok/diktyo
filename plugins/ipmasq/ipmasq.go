package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

type MasqEntry struct {
	External    string `json:"external,omitempty"`
	Destination string `json:"destination,omitempty"`
	Protocol    string `json:"protocol"`
	Description string `json:"description,omitempty"`
}

func (e MasqEntry) Valid() bool {
	if e.External == "" {
		return false
	}

	if e.Destination == "" {
		return false
	}
	return true
}

func (e MasqEntry) Insert(ip net.IP, chain string, handle string) error {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}
	if err := ensureChain(chain); err != nil {
		return err
	}
	if err := ensureChainInPlace(chain); err != nil {
		return err
	}
	comment := strings.Replace(fmt.Sprintf("%s: %s", handle, e.Description), " ", "_", -1)
	return ipt.AppendUnique("nat", chain, "-p", e.Protocol, "-d", e.Destination, "-s", ip.String(), "-j",
		"SNAT", "--to", e.External, "-m", "comment", "--comment", comment)
}

func (e MasqEntry) Delete(chain string, handle string) error {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}

	if err := ensureChain(chain); err != nil {
		return err
	}

	rules, err := ipt.List("nat", chain)
	if err != nil {
		return fmt.Errorf("unable to list the rules: %v", err)
	}

	for _, r := range rules {
		m, s := getArgumentMap(r)
		comment := strings.Replace(fmt.Sprintf("%s: %s", handle, e.Description), " ", "_", -1)
		if m["--comment"] == comment {
			delete(m, "-A")
			return ipt.Delete("nat", chain, s...)
		}
	}

	return nil
}

func ensureChain(name string) error {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}

	// Create chain if doesn't exist
	chains, err := ipt.ListChains("nat")
	if err != nil {
		return fmt.Errorf("failed to list chains: %v", err)
	}
	for _, ch := range chains {
		if ch == name {
			return nil
		}
	}
	return ipt.NewChain("nat", name)
}

func ensureChainInPlace(name string) error {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}

	return ipt.AppendUnique("nat", "POSTROUTING", "-j", name, "-m", "comment", "--comment", "ipmasq cni plugin")
}

//input:   -A mycoolChain -s 10.10.30.10/32  --comment "something_without_space"
//output:  map[string]string = { "-A": "mycoolChain", "-s": "10.10.30.10/32"}
func getArgumentMap(input string) (map[string]string, []string) {
	m := make(map[string]string)
	r := []string{}
	s := strings.Split(input, " ")
	for i := 0; i < len(s); i += 2 {
		if s[i] == "-A" {
			continue
		}
		m[s[i]] = strings.Replace(s[i+1], "\"", "", -1)
		r = append(r, s[i])
		r = append(r, m[s[i]])
	}
	return m, r
}
