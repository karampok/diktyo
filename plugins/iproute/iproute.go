package main

import (
	"fmt"
	"net"
	"runtime"

	"golang.org/x/sys/unix"

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
	Destination string `json:"destination,omitempty"`
	Gateway     string `json:"gateway,omitempty"`
	Description string `json:"description,omitempty"`
}

func setupRoutes(netns ns.NetNS, routes []RouteEntry) error {

	err := netns.Do(func(hostNS ns.NetNS) error {
		for _, r := range routes {
			_, dst, _ := net.ParseCIDR(r.Destination)
			route := netlink.Route{
				Dst: dst,
			}

			if r.Gateway == "drop" {
				route.Type = unix.RTN_BLACKHOLE
			} else {
				route.Gw = net.ParseIP(r.Gateway)
			}

			if err := netlink.RouteAdd(&route); err != nil {
				return fmt.Errorf("failed to add route %v: %v", dst, err)
			}
		}
		return nil
	})

	return err

}
