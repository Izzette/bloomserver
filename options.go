package main

import (
	"log"
	"net"
	"net/url"
	"strconv"
)

// Configuration options.
type Options struct {
	BloomFilterFile          string // Relative or absolute file path to the bloom filter file
	ListenAddress            string // Address to listen on
	HTTPMaxRequestBodyLength int    // Maximum HTTP request body length

	listenAddressNetwork string
	listenAddressIP      net.IP // If TCP
	listenAddressPort    uint16 // If TCP
	listenAddressPath    string // If UNIX domain socket
}

// Parses and validates configuration options from unprotected fields.
func (o *Options) Parse() {
	o.parseListenAddress()
	o.parseHTTPMaxReqestBodyLength()
}

// Provides network string compatible with `func net.Dial`.
func (o *Options) ListenAddressNetwork() string {
	return o.listenAddressNetwork
}

// Provides address string compatible with `func net.Dial`.
func (o *Options) ListenAddressAddress() string {
	switch o.listenAddressNetwork {
	case "tcp", "tcp4", "tcp6":
		ipStr := ""
		if o.listenAddressIP.To4() != nil {
			ipStr = o.listenAddressIP.String()
		} else if o.listenAddressIP.To16() != nil {
			ipStr = "[" + o.listenAddressIP.String() + "]"
		} else {
			log.Panic("IP address is neither 4-bytes nor 16-bytse?")
		}
		portStr := strconv.FormatUint(uint64(o.listenAddressPort), 10)
		return net.JoinHostPort(ipStr, portStr)
	case "unix":
		return o.listenAddressPath
	default:
		log.Panicf("Got strange value for listenAddressNetwork (%s): is this object initialized?", o.listenAddressNetwork)
		return ""
	}
}

func (o *Options) parseListenAddress() {
	var listenAddressURL *url.URL
	var err error

	if listenAddressURL, err = url.Parse(o.ListenAddress); err != nil {
		log.Fatalf("Could not parse ListenAddress (%s): %s\n", o.ListenAddress, err.Error())
	}

	o.listenAddressNetwork = listenAddressURL.Scheme

	switch o.listenAddressNetwork {
	case "tcp", "tcp4", "tcp6":
		var host, port string
		if host, port, err = net.SplitHostPort(listenAddressURL.Host); err != nil {
			log.Fatalf("Could not parse host and port from ListenAddress (%s): %s\n", o.ListenAddress, err.Error())
		}

		if o.listenAddressIP = net.ParseIP(host); o.listenAddressIP != nil {
			// nop
		} else if o.listenAddressIP, _ = lookupIPByNetwork(host, o.listenAddressNetwork); o.listenAddressIP != nil {
			// nop
		} else {
			log.Fatalf("Could not determine the appropriate IP address from ListenAddress (%s)\n", o.ListenAddress)
		}

		if portInt, err := strconv.ParseUint(port, 10, 16); err != nil {
			log.Fatalf("Could not parse the port from ListenAddress (%s): %s\n", o.ListenAddress, err)
		} else {
			o.listenAddressPort = uint16(portInt)
		}
	case "unix":
		o.listenAddressPath = listenAddressURL.Path
	default:
		log.Fatalf("Unsupported protocol (%s) for ListenAddress (%s)!\n", o.listenAddressNetwork, o.ListenAddress)
	}
}

func (o *Options) parseHTTPMaxReqestBodyLength() {
	if o.HTTPMaxRequestBodyLength <= 0 {
		log.Fatalf("Maximum request body length is out-of-bounds, cannot be zero or negative")
	}
}

func lookupIPByNetwork(host string, network string) (net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, nil
	}

	switch network {
	case "tcp":
		// The first is good-enough
		return ips[0], nil
	case "tcp4":
		for _, ip := range ips {
			if ip.To4() != nil {
				return ip, nil
			}
		}
	case "tcp6":
		for _, ip := range ips {
			if ip.To16() != nil {
				return ip, nil
			}
		}
		return nil, nil
	}

	return nil, nil
}
