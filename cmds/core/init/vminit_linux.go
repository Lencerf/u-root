package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

const (
	ula1Name = "ula1"
	ula1Addr = "fd14:988a:50ee:10d::1/126"
	ula1Mac  = "02:32:10:d0:00:01"

	eth0AddrPrefix = "192.168.14."
	eth0MacPrefix  = "ea:c2:14:80:10:"
	gatewayIP      = "192.168.14.254"
)

var (
	oakDice       = flag.Int("oak-dice", 0, "Oak dice")
	oakEventLog   = flag.Int("oak-event-log", 0x13000, "Oak event log")
	oakDiceLength = flag.Int("oak-dice-length", 12288, "Oak dice length")
)

func netInit() error {
	links, err := netlink.LinkList()
	if err != nil {
		return err
	}
	for _, l := range links {
		macAddr := l.Attrs().HardwareAddr.String()
		log.Printf("Find link %s: %s...", l.Attrs().Name, macAddr)
		if strings.HasPrefix(macAddr, eth0MacPrefix) {
			index, err := strconv.ParseUint(strings.TrimPrefix(macAddr, eth0MacPrefix), 16, 8)
			if err != nil {
				return fmt.Errorf("failed to parse index: %w", err)
			}
			addr, err := netlink.ParseAddr(fmt.Sprintf("%s%d/24", eth0AddrPrefix, index))
			if err != nil {
				return fmt.Errorf("failed to parse address: %w", err)
			}
			if err := netlink.AddrAdd(l, addr); err != nil {
				return fmt.Errorf("failed to add address: %w", err)
			}
			if err := netlink.LinkSetUp(l); err != nil {
				return fmt.Errorf("failed to set link up: %w", err)
			}
			route := netlink.Route{LinkIndex: l.Attrs().Index, Gw: net.ParseIP(gatewayIP)}
			if err := netlink.RouteAdd(&route); err != nil {
				return fmt.Errorf("failed to add route: %w", err)
			}
			log.Printf("Done configuring interface %s", l.Attrs().Name)

		} else if macAddr == ula1Mac {
			if err := netlink.LinkSetName(l, ula1Name); err != nil {
				return fmt.Errorf("failed to rename link: %w", err)
			}
			addr, err := netlink.ParseAddr(ula1Addr)
			if err != nil {
				return fmt.Errorf("failed to parse address: %w", err)
			}
			if err := netlink.AddrAdd(l, addr); err != nil {
				return fmt.Errorf("failed to add address: %w", err)
			}
			if err := netlink.LinkSetUp(l); err != nil {
				return fmt.Errorf("failed to set link up: %w", err)
			}
			log.Println("Done configuring interface ula1")
		}
	}
	log.Println("Done net init")
	return nil
}

func fsInit() error {
	os.Setenv("HISTFILE", "/root/.bash_history")

	os.MkdirAll("/data", 0o700)
	unix.Mount("data", "/data", "virtiofs", 0, "")
	unix.Mount("data-dax", "/data", "virtiofs", 0, "dax=always")

	return nil
}
