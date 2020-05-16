package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/Izzette/bloomserver/bloom"
)

func main() {
	bloomFilterFile := flag.String("bloom-filter-file", "", "path to the bloom filter file")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Panic("Must specify an action")
	}

	if *bloomFilterFile == "" {
		log.Fatal("Must provide a value for -bloom-filter-file")
	}

	switch args[0] {
	case "new":
		if len(args) != 3 {
			log.Fatal("Must provide exactly two arguments to the `new` action.")
		}

		m, err := strconv.ParseUint(args[1], 10, 0)
		if err != nil {
			log.Fatalf("Invalid value for bits size (M): %s\n", err.Error())
		}
		k, err := strconv.ParseUint(args[2], 10, 0)
		if err != nil {
			log.Fatalf("Invalid value for hash functions (K): %s\n", err.Error())
		}

		f := bloom.New(uint(m), uint(k))

		saveToFile(f, *bloomFilterFile)
	case "add":
		if len(args) <= 1 {
			log.Fatal("Must provide at least one word to add.")
		}

		f := bloom.FromFile(*bloomFilterFile)
		for _, word := range args[1:] {
			f.Filter.Add([]byte(word))
		}

		saveToFile(f, *bloomFilterFile)
	default:
		log.Fatalf("Unknown action: %s\n", args[0])
	}
}

func saveToFile(f *bloom.BloomFilter, filePath string) {
	if file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err != nil {
		log.Panicf("Could not create new bloom filter file (%s): %s\n", filePath, err.Error())
	} else {
		defer file.Close()
		f.Save(file)
	}
}
