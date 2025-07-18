package bloomfilter

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"

	"github.com/another-d-mention/unicomplex/crypt/hashing"
	"github.com/another-d-mention/unicomplex/datastruct/bitset"
)

type BloomFilter struct {
	data *bitset.BitSet
	m    uint
	k    uint
}

func EstimateParameters(n uint, p float64) (m uint, k uint) {
	m = uint(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
	k = uint(math.Ceil(math.Log(2) * float64(m) / float64(n)))
	return
}

// EstimateFalsePositiveRate returns, for a BloomFilter of m bits
// and k hash functions, an estimation of the false positive rate when
//
//	storing n entries. This is an empirical, relatively slow
//
// test using integers as keys.
// This function is useful to validate the implementation.
func EstimateFalsePositiveRate(m, k, n uint) (fpRate float64) {
	rounds := uint32(100000)
	// We construct a new filter.
	f := New(m, k)
	n1 := make([]byte, 4)
	// We populate the filter with n values.
	for i := uint32(0); i < uint32(n); i++ {
		binary.BigEndian.PutUint32(n1, i)
		f.Add(n1)
	}
	fp := 0
	// test for number of rounds
	for i := uint32(0); i < rounds; i++ {
		binary.BigEndian.PutUint32(n1, i+uint32(n)+1)
		if f.Test(n1) {
			fp++
		}
	}
	fpRate = float64(fp) / (float64(rounds))
	return
}

// New creates a new Bloom filter with the given number of bits and hashCount hashing functions
func New(bitCount, hashCount uint) *BloomFilter {
	if bitCount < 1 {
		bitCount = 1
	}
	if hashCount < 1 {
		hashCount = 1
	}
	return &BloomFilter{
		data: bitset.New(bitCount),
		m:    bitCount,
		k:    hashCount,
	}
}

// Cap returns the capacity, _m_, of a Bloom filter
func (f *BloomFilter) Cap() uint {
	return f.m
}

// K returns the number of hash functions used in the BloomFilter
func (f *BloomFilter) K() uint {
	return f.k
}

// Add data to the Bloom Filter. Returns the filter (allows chaining)
func (f *BloomFilter) Add(data []byte) *BloomFilter {
	h := baseHashes(data)
	for i := uint(0); i < f.k; i++ {
		f.data.Set(f.location(h, i))
	}
	return f
}

// Test returns true if the data is in the BloomFilter, false otherwise.
// If true, the result might be a false positive. If false, the data
// is definitely not in the set.
func (f *BloomFilter) Test(data []byte) bool {
	h := baseHashes(data)
	for i := uint(0); i < f.k; i++ {
		if !f.data.Test(f.location(h, i)) {
			return false
		}
	}
	return true
}

func (f *BloomFilter) Reset() {
	f.data.ClearAll()
}

// ApproximatedSize approximates the number of items
func (f *BloomFilter) ApproximatedSize() uint32 {
	x := float64(f.data.Count())
	m := float64(f.Cap())
	k := float64(f.K())
	size := -1 * m / k * math.Log(1-x/m) / math.Log(math.E)
	return uint32(math.Floor(size + 0.5)) // round
}

// location returns the ith hashed location using the four base hash values
func (f *BloomFilter) location(h [4]uint64, i uint) uint {
	return uint(location(h, i) % uint64(f.m))
}

// bloomFilterJSON is an unexported type for marshaling/unmarshaling BloomFilter struct.
type bloomFilterJSON struct {
	M uint           `json:"m"`
	K uint           `json:"k"`
	B *bitset.BitSet `json:"b"`
}

// MarshalJSON implements json.Marshaler interface.
func (f *BloomFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(bloomFilterJSON{f.m, f.k, f.data})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (f *BloomFilter) UnmarshalJSON(data []byte) error {
	var j bloomFilterJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		return err
	}
	f.m = j.M
	f.k = j.K
	f.data = j.B
	return nil
}

// GobEncode implements gob.GobEncoder interface.
func (f *BloomFilter) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	_, err := f.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GobDecode implements gob.GobDecoder interface.
func (f *BloomFilter) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	_, err := f.ReadFrom(buf)

	return err
}

// MarshalBinary implements binary.BinaryMarshaler interface.
func (f *BloomFilter) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	_, err := f.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements binary.BinaryUnmarshaler interface.
func (f *BloomFilter) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	_, err := f.ReadFrom(buf)

	return err
}

// WriteTo writes a binary representation of the BloomFilter to an i/o stream.
// It returns the number of bytes written.
//
// Performance: if this function is used to write to a disk or network
// connection, it might be beneficial to wrap the stream in a bufio.Writer.
// E.g.,
//
//	      f, err := os.Create("myfile")
//		       w := bufio.NewWriter(f)
func (f *BloomFilter) WriteTo(stream io.Writer) (int64, error) {
	err := binary.Write(stream, binary.BigEndian, uint64(f.m))
	if err != nil {
		return 0, err
	}
	err = binary.Write(stream, binary.BigEndian, uint64(f.k))
	if err != nil {
		return 0, err
	}
	numBytes, err := f.data.WriteTo(stream)
	return numBytes + int64(2*binary.Size(uint64(0))), err
}

// ReadFrom reads a binary representation of the BloomFilter (such as might
// have been written by WriteTo()) from an i/o stream. It returns the number
// of bytes read.
//
// Performance: if this function is used to read from a disk or network
// connection, it might be beneficial to wrap the stream in a bufio.Reader.
// E.g.,
//
//	f, err := os.Open("myfile")
//	r := bufio.NewReader(f)
func (f *BloomFilter) ReadFrom(stream io.Reader) (int64, error) {
	var m, k uint64
	err := binary.Read(stream, binary.BigEndian, &m)
	if err != nil {
		return 0, err
	}
	err = binary.Read(stream, binary.BigEndian, &k)
	if err != nil {
		return 0, err
	}
	b := &bitset.BitSet{}
	numBytes, err := b.ReadFrom(stream)
	if err != nil {
		return 0, err
	}
	f.m = uint(m)
	f.k = uint(k)
	f.data = b
	return numBytes + int64(2*binary.Size(uint64(0))), nil
}

// baseHashes returns the four hash values of data that are used to create k
// hashes
func baseHashes(data []byte) [4]uint64 {
	sum := hashing.NewMurmur256Hasher().Bytes(data)
	hash1 := binary.BigEndian.Uint64(sum[0:])
	hash2 := binary.BigEndian.Uint64(sum[8:])
	hash3 := binary.BigEndian.Uint64(sum[16:])
	hash4 := binary.BigEndian.Uint64(sum[24:])
	return [4]uint64{
		hash1, hash2, hash3, hash4,
	}
}

// location returns the ith hashed location using the four base hash values
func location(h [4]uint64, i uint) uint64 {
	ii := uint64(i)
	return h[ii%2] + ii*h[2+(((ii+(ii%2))%4)/2)]
}
