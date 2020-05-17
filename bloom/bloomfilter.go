package bloom

import (
	"encoding/binary"
	"log"
	"os"

	"github.com/willf/bloom"
)

type BloomFilter struct {
	Filter *bloom.BloomFilter

	k uint
	m uint
}

const BLOOM_FILTER_MAGIC = "\x01IZZETTE/BLOOMSERVER\x03\x02"

func New(m uint, k uint) (f *BloomFilter) {
	f = &BloomFilter{
		Filter: bloom.New(m, k),

		m: m,
		k: k,
	}

	return f
}

func FromFile(filterFilePath string) (f *BloomFilter) {
	f = &BloomFilter{}
	if file, err := os.OpenFile(filterFilePath, os.O_RDONLY, 0); err != nil {
		log.Panicf("Failed to open filter file (%s): %s\n", filterFilePath, err.Error())
	} else {
		defer file.Close()
		f.m, f.k = parseFilterFile(file)
		f.Filter = bloom.New(f.m, f.k)
		f.Filter.ReadFrom(file)
	}

	return f
}

func (f *BloomFilter) Save(file *os.File) {
	if _, err := file.Write([]byte(BLOOM_FILTER_MAGIC)); err != nil {
		log.Panicf("Couldn't to write magic bytes to filter file (%s): %s\n", file.Name(), err.Error())
	}

	kBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(kBytes, uint64(f.k))
	if _, err := file.Write(kBytes); err != nil {
		log.Panicf("Couldn't to write K to filter file (%s): %s\n", file.Name(), err.Error())
	}

	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(f.m))
	if _, err := file.Write(sizeBytes); err != nil {
		log.Panicf("Couldn't to write size to filter file (%s): %s\n", file.Name(), err.Error())
	}

	f.Filter.WriteTo(file)
}

func (f *BloomFilter) GetM() uint {
	return f.m
}

func (f *BloomFilter) GetK() uint {
	return f.k
}

func parseFilterFile(file *os.File) (size uint, k uint) {
	magicBytes := []byte(BLOOM_FILTER_MAGIC)
	headerBytes := make([]byte, len(magicBytes))
	if _, err := file.Read(headerBytes); err != nil {
		log.Panicf("Couldn't read magic bytes in filter file (%s): %s\n", file.Name(), err.Error())
	} else if string(headerBytes) != BLOOM_FILTER_MAGIC {
		log.Fatalf("Couldn't find magic bytes in filter file (%s): no match, is this a filter file?", file.Name())
	}

	kBytes := make([]byte, 8)
	if _, err := file.Read(kBytes); err != nil {
		log.Panicf("Couldn't read K from filter file (%s): %s\n", file.Name(), err.Error())
	}
	k = uint(binary.LittleEndian.Uint64(kBytes))

	sizeBytes := make([]byte, 8)
	if _, err := file.Read(sizeBytes); err != nil {
		log.Panicf("Couldn't read size of bloom filter in filter file (%s): %s\n", file.Name(), err.Error())
	}
	size = uint(binary.LittleEndian.Uint64(sizeBytes))

	return size, k
}
