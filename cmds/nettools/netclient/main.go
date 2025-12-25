package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/vsock"
)

var (
	addr    = flag.String("addr", "", "address")
	network = flag.String("net", "tcp", "net: tcp, unix, vsock, unix-vsock")
	delay   = flag.Duration("delay", 0, "delay")
)

func app() (retErr error) {
	var c net.Conn
	var err error
	switch *network {
	case "tcp", "unix":
		c, err = net.Dial(*network, *addr)
	case "unix-vsock":
		parts := strings.SplitN(*addr, ":", 2)
		uds := parts[0]
		port, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return err
		}
		c, err = net.Dial("unix", uds)
		if err != nil {
			return err
		}
		connectMsg := fmt.Sprintf("CONNECT %d\n", port)
		if _, err := io.WriteString(c, connectMsg); err != nil {
			return fmt.Errorf("sending connect request: %w", err)
		}
		buf := make([]byte, 2)
		if _, err := io.ReadFull(c, buf); err != nil {
			return fmt.Errorf("reading connect request: %w", err)
		}
		if string(buf) != "OK" {
			return fmt.Errorf("vsock: expect OK, got %s", buf)
		}
		buf = make([]byte, 1)
		for buf[0] != '\n' {
			if _, err := io.ReadFull(c, buf); err != nil {
				return err
			}
		}
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
		fmt.Printf("dial: %d: %d\n", cid, port)
		c, err = vsock.Dial(uint32(cid), uint32(port), nil)
	default:
		return fmt.Errorf("unsupported network: %s", *network)
	}
	if err != nil {
		return err
	}
	defer func() {
		if e := c.Close(); e != nil && retErr == nil {
			retErr = e
		}
	}()
	buf := make([]byte, 128)
	n, err := c.Read(buf)
	if err != nil {
		return err
	}
	fmt.Printf("resp: %q\n", buf[:n])
	if *delay > 0 {
		time.Sleep(*delay)
	}
	return nil
}

func main() {
	flag.Parse()
	if err := app(); err != nil {
		log.Fatal(err)
	}
}
