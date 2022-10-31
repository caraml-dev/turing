package compression

import (
	"github.com/pierrec/lz4"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
)

// Compressor is an interface to be implemented by types that support reading and writing
// of compressed data of a specific format
type Compressor interface {
	Compress([]byte) ([]byte, error)
	Uncompress([]byte) ([]byte, error)
}

// LZ4Compressor uses the lz4 algorithm to convert byte arrays
type LZ4Compressor struct{}

// Compress takes an input byte array and compresses it using the LZ4 algorithm
func (l *LZ4Compressor) Compress(data []byte) ([]byte, error) {
	// Allocate 64K for hash table
	ht := make([]int, 64<<10)
	compressed := make([]byte, len(data))

	n, err := lz4.CompressBlock(data, compressed, ht)
	if err != nil {
		return nil, errors.Wrapf(err, "LZ4 compression failed")
	}

	return compressed[:n], nil
}

// Uncompress takes an input byte array and uncompresses it using the LZ4 algorithm
func (l *LZ4Compressor) Uncompress(data []byte) ([]byte, error) {
	// Arbitrarily allocate 10 times the compressed data length for uncompress
	uncompressed := make([]byte, 10*len(data))

	n, err := lz4.UncompressBlock(data, uncompressed)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, "LZ4 uncompression failed: %s", err.Error())
	}

	return uncompressed[:n], nil
}
