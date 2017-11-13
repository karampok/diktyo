package main

import (
	"fmt"
	"net"

	"github.com/coreos/go-iptables/iptables"
)

//TODO.
var chainName = "MyDefaultMasqRuleChain"

//iptables -t nat -A POSTROUTING -p tcp -o eth0 -j SNAT --to 1.2.3.4:1-1023
type MasqEntry struct {
	External string `json:"external,omitempty"`
	Protocol string `json:"protocol"`
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

	return ipt.AppendUnique("nat", "POSTROUTING", "-j", name, "-m", "comment", "--comment", "diktyo")
}

func (e MasqEntry) Valid() bool {
	if e.External == "" {
		return false
	}
	return true
}

func (e MasqEntry) Apply(ip net.IP, chain string, comment string) error {
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

	return ipt.AppendUnique("nat", chain, "-p", e.Protocol, "-s", ip.String(), "-j", "SNAT", "--to", e.External, "-m", "comment", "--comment", comment)
}

func (e MasqEntry) Delete(ip net.IP, chain string, comment string) error {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}

	return ipt.Delete("nat", chain, "-p", e.Protocol, "-s", ip.String(), "-j", "SNAT", "--to", e.External, "-m", "comment", "--comment", comment)
}
