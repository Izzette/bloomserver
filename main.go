package main

import (
	"flag"
	"log"
)

func main() {
	bloomFilterFile := flag.String("bloom-filter-file", "", "path to the bloom filter file")
	listenAddress := flag.String("listen-address", "tcp://127.0.0.1:14519", "address to listen on (default tcp://127.0.0.1:14519)")
	httpMaxRequestBodyLength := flag.Int("http-max-request-body-length", 4096, "the maximium HTTP request body length (default 4KiB)")

	flag.Parse()

	o := Options{
		ListenAddress:            *listenAddress,
		BloomFilterFile:          *bloomFilterFile,
		HTTPMaxRequestBodyLength: *httpMaxRequestBodyLength,
	}
	o.Parse()

	s := NewBloomHTTPServer(o)
	s.Start()
	if err := <-s.Errors; err != nil {
		log.Panicf("Error while running HTTP server: %s\n", err.Error())
	}
}
