package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/Izzette/bloomserver/bloom"
)

type BloomHTTPServer struct {
	BloomFilter *bloom.BloomFilter
	Options     Options
	Errors      chan error
	Server      *http.Server
}

// Handles GET to /api/search
func (s BloomHTTPServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	ctx, ctxCancel := context.WithDeadline(r.Context(), time.Now().Add(5*time.Second))
	defer ctxCancel()
	r = r.WithContext(ctx)

	var substringLength int
	if s := r.URL.Query().Get("substringLength"); s == "" {
		substringLength = 0
	} else if l, err := strconv.ParseUint(s, 10, 16); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		substringLength = int(l)
	}

	if r.ContentLength <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.ContentLength > int64(s.Options.HTTPMaxRequestBodyLength) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	bodyData := make([]byte, r.ContentLength)
	if n, err := r.Body.Read(bodyData); int64(n) != r.ContentLength && err != nil {
		log.Printf("ERROR: Got something strange while reading request data after reading %d bytes: %s\n", n, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if int64(n) != r.ContentLength {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !utf8.Valid(bodyData) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body := []rune(string(bodyData))

	guiltySubstrings := make([]string, 0)
	if substringLength == 0 {
		substringLength = len(body)
	}

	// Iterate over all substrings
	for i := 0; len(body) >= i+substringLength; i += 1 {
		for j := substringLength; len(body) >= i+j; j += 1 {
			// Check if deadline exceeded
			if _, ok := r.Context().Deadline(); !ok {
				log.Print(r.Context().Deadline())
				w.WriteHeader(http.StatusRequestTimeout)
				return
			}

			substring := string(body[i : i+j])
			if s.BloomFilter.Filter.Test([]byte(substring)) {
				guiltySubstrings = append(guiltySubstrings, substring)
			}
		}
	}

	responseObject := map[string]interface{}{
		"guiltySubstrings": guiltySubstrings,
	}
	if jsonData, err := json.Marshal(responseObject); err != nil {
		log.Printf("Error encoding response object as JSON: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func NewBloomHTTPServer(o Options) (s *BloomHTTPServer) {
	s = &BloomHTTPServer{
		BloomFilter: bloom.FromFile(o.BloomFilterFile),
		Options:     o,
		Errors:      make(chan error, 1),
	}

	muxHandler := http.NewServeMux()
	muxHandler.HandleFunc("/api/search", s.handleSearch)

	s.Server = &http.Server{
		Handler: muxHandler,
	}

	return s
}

func (s *BloomHTTPServer) Start() {
	if s.Options.ListenAddressNetwork() == "unix" {
		// We don't really care about the error, if we didn't succeed in removing, we should just try to listen instead.
		_ = os.Remove(s.Options.ListenAddressAddress())
	}
	listener, err := net.Listen(s.Options.ListenAddressNetwork(), s.Options.ListenAddressAddress())
	if err != nil {
		log.Panicf("Could not bind to socket (%s): %s\n", s.Options.ListenAddress, err.Error())
	}
	log.Printf("Listening on %s (%s) ...\n", s.Options.ListenAddressAddress(), s.Options.ListenAddressNetwork())

	go func() {
		defer listener.Close()
		s.Errors <- s.Server.Serve(listener)
	}()
}
