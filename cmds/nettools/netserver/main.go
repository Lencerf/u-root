package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/mdlayher/vsock"
)

var (
	addr    = flag.String("addr", "", "uds path")
	network = flag.String("net", "tcp", "net")
	msg     = flag.String("msg", "Hello", "uds message")
	num     = flag.Int("num", 1, "number of services")
)

func app() error {
	var listener net.Listener
	var err error
	switch *network {
	case "tcp", "unix":
		listener, err = net.Listen(*network, *addr)
	case "vsock":
		parts := strings.Split(*addr, ":")
		cid := uint64(0)
		port := uint64(0)
		cid, err = strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return err
		}
		port, err = strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return err
		}
		listener, err = vsock.ListenContextID(uint32(cid), uint32(port), nil)
	default:
		return fmt.Errorf("unsupported network: %s", *network)
	}
	if err != nil {
		return err
	}
	for i := 0; *num < 0 || i < *num; i++ {
		client, err := listener.Accept()
		if err != nil {
			log.Printf("accept: %s", err)
			continue
		}
		if n, err := client.Write([]byte(*msg)); err != nil || n != len(*msg) {
			log.Printf("write %s, wrote %d bytes, err = %s", *msg, n, err)
			continue
		}
		if err := client.Close(); err != nil {
			log.Printf("close: %s", err)
		}
	}
	return listener.Close()
}

func main() {
	flag.Parse()
	if err := app(); err != nil {
		log.Fatal(err)
	}
}
