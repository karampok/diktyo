package main

import (
	"fmt"
	"net"
	"runtime"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

type RouteEntry struct {
	CDIR        string `json:"external,omitempty"`
	Gateway     string `json:"destination,omitempty"`
	Description string `json:"description,omitempty"`
}

func setupRoutes(netns ns.NetNS, routes []RouteEntry) error {

	err := netns.Do(func(hostNS ns.NetNS) error {
		for _, r := range routes {
			route := netlink.Route{
				Dst: &net.IPNet{
					IP:   net.ParseIP("192.0.2.1"),
					Mask: net.CIDRMask(31, 32),
				},
				Scope: netlink.SCOPE_NOWHERE,
			}

			if err := netlink.RouteAdd(&route); err != nil {
				return fmt.Errorf("failed to add route %v: %v", r, err)
			}
		}
		return nil
	})

	return err

}
