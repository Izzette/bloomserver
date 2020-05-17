package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Izzette/bloomserver/bloom"

	"code.cloudfoundry.org/bytefmt"
	willf_bloom "github.com/willf/bloom"
)

func main() {
	bloomFilterFile := flag.String("bloom-filter-file", "", "path to the bloom filter file")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Panic("Must specify an action")
	}

	switch args[0] {
	case "create":
		if len(args) != 3 {
			log.Fatal("Must provide exactly two arguments to the `create` action.")
		}

		if *bloomFilterFile == "" {
			log.Fatal("Must provide a value for -bloom-filter-file")
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
		if len(args) != 2 {
			log.Fatal("Must provide exactly two arguments to the `add` action.")
		}

		if *bloomFilterFile == "" {
			log.Fatal("Must provide a value for -bloom-filter-file")
		}

		f := bloom.FromFile(*bloomFilterFile)

		var scanner *bufio.Scanner
		if wordFile, err := os.OpenFile(args[1], os.O_RDONLY, 0); err != nil {
			log.Panicf("Failed to open word-list file (%s): %s\n", args[1], err.Error())
		} else {
			defer wordFile.Close()
			scanner = bufio.NewScanner(bufio.NewReader(wordFile))
		}

		for scanner.Scan() {
			word := scanner.Text()

			f.Filter.Add([]byte(word))
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("Encountered error while scanning word-list file (%s): %s\n", args[1], err.Error())
		}

		saveToFile(f, *bloomFilterFile)
	case "estimate":
		if len(args) != 3 {
			log.Fatal("Must provide exactly two arguments to the `estimate` action.")
		}

		var n uint
		if n64, err := strconv.ParseUint(args[1], 10, 0); err != nil {
			log.Fatalf("Couldn't parse estimated number of entries (%s): %s\n", args[1], err.Error())
		} else {
			n = uint(n64)
		}

		var p float64
		if p64, err := strconv.ParseFloat(args[2], 64); err != nil {
			log.Fatalf("Couldn't parse acceptable false-positive ratio (%s): %s\n", args[2], err.Error())
		} else {
			p = p64
		}

		m, k := willf_bloom.EstimateParameters(n, p)

		mBytes := m / 8
		if m%8 != 0 {
			mBytes += 1
		}
		mBytesInUint64 := 8 * (mBytes / 8)
		if mBytes%8 != 0 {
			mBytesInUint64 += 8
		}

		estimatedBytesSize := uint64(mBytesInUint64) + uint64(len(bloom.BLOOM_FILTER_MAGIC)) + 16

		fmt.Printf("M (number of bits): %d (~%siB), K (number of hash functions): %d\n", m, bytefmt.ByteSize(estimatedBytesSize), k)
	case "show":
		if len(args) != 1 {
			log.Fatal("The `show` action accepts zero arguments.")
		}

		if *bloomFilterFile == "" {
			log.Fatal("Must provide a value for -bloom-filter-file")
		}

		f := bloom.FromFile(*bloomFilterFile)
		m := f.GetM()
		k := f.GetK()

		fmt.Printf("%s: M (number of bits): %d, K (number of hash functions): %d\n", *bloomFilterFile, m, k)
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
