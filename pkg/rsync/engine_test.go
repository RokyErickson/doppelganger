package rsync

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestBlockHashNilInvalid(t *testing.T) {
	var hash *BlockHash
	if hash.EnsureValid() == nil {
		t.Error("nil block hash considered valid")
	}
}

func TestBlockHashNilStrongHashInvalid(t *testing.T) {
	hash := &BlockHash{Weak: 5}
	if hash.EnsureValid() == nil {
		t.Error("block hash with nil strong hash considered valid")
	}
}

func TestBlockHashEmptyStrongHashInvalid(t *testing.T) {
	hash := &BlockHash{Weak: 5, Strong: make([]byte, 0)}
	if hash.EnsureValid() == nil {
		t.Error("block hash with empty strong hash considered valid")
	}
}

func TestSignatureNilInvalid(t *testing.T) {
	var signature *Signature
	if signature.EnsureValid() == nil {
		t.Error("nil signature considered valid")
	}
}

func TestSignatureZeroBlockSizeNonZeroLastBlockSizeInvalid(t *testing.T) {
	signature := &Signature{LastBlockSize: 8192}
	if signature.EnsureValid() == nil {
		t.Error("zero block size with non-zero last block size considered valid")
	}
}

func TestSignatureZeroBlockSizeWithHashesInvalid(t *testing.T) {
	signature := &Signature{Hashes: []*BlockHash{{Weak: 5, Strong: []byte{0x0}}}}
	if signature.EnsureValid() == nil {
		t.Error("zero block size with hashes considered valid")
	}
}

func TestSignatureZeroLastBlockSizeInvalid(t *testing.T) {
	signature := &Signature{BlockSize: 8192}
	if signature.EnsureValid() == nil {
		t.Error("zero last block size considered valid")
	}
}

func TestSignatureLastBlockSizeTooBigInvalid(t *testing.T) {
	signature := &Signature{BlockSize: 8192, LastBlockSize: 8193}
	if signature.EnsureValid() == nil {
		t.Error("overly large last block size considered valid")
	}
}

func TestSignatureNoHashesInvalid(t *testing.T) {
	signature := &Signature{BlockSize: 8192, LastBlockSize: 8192}
	if signature.EnsureValid() == nil {
		t.Error("signature with no hashes considered valid")
	}
}

func TestSignatureInvalidHashesInvalid(t *testing.T) {
	signature := &Signature{
		BlockSize:     8192,
		LastBlockSize: 8192,
		Hashes:        []*BlockHash{nil},
	}
	if signature.EnsureValid() == nil {
		t.Error("signature with no hashes considered valid")
	}
}

func TestSignatureValid(t *testing.T) {
	signature := &Signature{
		BlockSize:     8192,
		LastBlockSize: 8192,
		Hashes:        []*BlockHash{{Weak: 1, Strong: []byte{0x0}}},
	}
	if err := signature.EnsureValid(); err != nil {
		t.Error("valid signature failed validation:", err)
	}
}

func TestOperationNilInvalid(t *testing.T) {
	var operation *Operation
	if operation.EnsureValid() == nil {
		t.Error("nil operation considered valid")
	}
}

func TestOperationDataAndStartInvalid(t *testing.T) {
	operation := &Operation{Data: []byte{0}, Start: 4}
	if operation.EnsureValid() == nil {
		t.Error("operation with data and start considered valid")
	}
}

func TestOperationDataAndCountInvalid(t *testing.T) {
	operation := &Operation{Data: []byte{0}, Count: 4}
	if operation.EnsureValid() == nil {
		t.Error("operation with data and count considered valid")
	}
}

func TestOperationZeroCountInvalid(t *testing.T) {
	operation := &Operation{Start: 40}
	if operation.EnsureValid() == nil {
		t.Error("operation with zero count considered valid")
	}
}

func TestOperationDataValid(t *testing.T) {
	operation := &Operation{Data: []byte{0}}
	if err := operation.EnsureValid(); err != nil {
		t.Error("valid data operation considered invalid")
	}
}

func TestOperationBlocksValid(t *testing.T) {
	operation := &Operation{Start: 10, Count: 50}
	if err := operation.EnsureValid(); err != nil {
		t.Error("valid block operation considered invalid")
	}
}

func TestMinimumBlockSize(t *testing.T) {
	if s := OptimalBlockSizeForBaseLength(1); s != minimumOptimalBlockSize {
		t.Error("incorrect minimum block size:", s, "!=", minimumOptimalBlockSize)
	}
}

func TestMaximumBlockSize(t *testing.T) {
	if s := OptimalBlockSizeForBaseLength(maximumOptimalBlockSize * maximumOptimalBlockSize); s != maximumOptimalBlockSize {
		t.Error("incorrect maximum block size:", s, "!=", maximumOptimalBlockSize)
	}
}

func TestOptimalBlockSizeForBase(t *testing.T) {

	baseLength := uint64(1234567)
	base := bytes.NewReader(make([]byte, baseLength))

	optimalBlockSize, err := OptimalBlockSizeForBase(base)
	if err != nil {
		t.Fatal("unable to compute optimal block size for base")
	}

	expectedOptimalBlockSize := OptimalBlockSizeForBaseLength(baseLength)
	if optimalBlockSize != expectedOptimalBlockSize {
		t.Error(
			"mismatch between optimal block size computations:",
			optimalBlockSize, "!=", expectedOptimalBlockSize,
		)
	}

	if uint64(base.Len()) != baseLength {
		t.Error("base was not reset to beginning")
	}
}

type testDataGenerator struct {
	length    int
	seed      int64
	mutations []int
	prepend   []byte
}

func (g testDataGenerator) generate() []byte {

	random := rand.New(rand.NewSource(g.seed))

	result := make([]byte, g.length)
	random.Read(result)

	for _, index := range g.mutations {
		result[index] += 1
	}

	if len(g.prepend) > 0 {
		result = append(g.prepend, result...)
	}

	return result
}

type engineTestCase struct {
	base                      testDataGenerator
	target                    testDataGenerator
	blockSize                 uint64
	maxDataOpSize             uint64
	numberOfOperations        uint
	numberOfDataOperations    uint
	expectCoalescedOperations bool
}

func (c engineTestCase) run(t *testing.T) {

	t.Helper()
	base := c.base.generate()
	target := c.target.generate()
	engine := NewEngine()
	signature := engine.BytesSignature(base, c.blockSize)
	if err := signature.EnsureValid(); err != nil {
		t.Fatal("generated signature was invalid:", err)
	} else if len(signature.Hashes) != 0 {
		if c.blockSize != 0 && signature.BlockSize != c.blockSize {
			t.Error(
				"generated signature did not have correct block size:",
				signature.BlockSize, "!=", c.blockSize,
			)
		}
	}

	delta := engine.DeltafyBytes(target, signature, c.maxDataOpSize)

	expectedMaxDataOpSize := c.maxDataOpSize
	if expectedMaxDataOpSize == 0 {
		expectedMaxDataOpSize = DefaultMaximumDataOperationSize
	}

	nDataOperations := uint(0)
	haveCoalescedOperations := false
	for _, o := range delta {
		if err := o.EnsureValid(); err != nil {
			t.Error("invalid operation:", err)
		} else if dataLength := uint64(len(o.Data)); dataLength > 0 {
			if dataLength > expectedMaxDataOpSize {
				t.Error(
					"data operation size greater than allowed:",
					dataLength, ">", expectedMaxDataOpSize,
				)
			}
			nDataOperations += 1
		} else if o.Count > 1 {
			haveCoalescedOperations = true
		}
	}
	if uint(len(delta)) != c.numberOfOperations {
		t.Error(
			"observed different number of operations than expected:",
			len(delta), "!=", c.numberOfOperations,
		)
	}
	if nDataOperations != c.numberOfDataOperations {
		t.Error(
			"observed different number of data operations than expected:",
			nDataOperations, ">", c.numberOfDataOperations,
		)
	}
	if haveCoalescedOperations != c.expectCoalescedOperations {
		t.Error(
			"expectations about coalescing not met:",
			haveCoalescedOperations, "!=", c.expectCoalescedOperations,
		)
	}

	patched, err := engine.PatchBytes(base, signature, delta)
	if err != nil {
		t.Fatal("unable to patch bytes:", err)
	}

	if !bytes.Equal(patched, target) {
		t.Error("patched data did not match expected")
	}
}

func TestBothEmpty(t *testing.T) {
	test := engineTestCase{
		base:   testDataGenerator{},
		target: testDataGenerator{},
	}
	test.run(t)
}

func TestBaseEmptyMaxDataOperationMultiple(t *testing.T) {
	test := engineTestCase{
		base:                   testDataGenerator{},
		target:                 testDataGenerator{10240, 473, nil, nil},
		maxDataOpSize:          1024,
		numberOfOperations:     10,
		numberOfDataOperations: 10,
	}
	test.run(t)
}

func TestBaseEmptyNonMaxDataOperationMultiple(t *testing.T) {
	test := engineTestCase{
		base:                   testDataGenerator{},
		target:                 testDataGenerator{10241, 473, nil, nil},
		maxDataOpSize:          1024,
		numberOfOperations:     11,
		numberOfDataOperations: 11,
	}
	test.run(t)
}

func TestTargetEmpty(t *testing.T) {
	test := engineTestCase{
		base:   testDataGenerator{12345, 473, nil, nil},
		target: testDataGenerator{},
	}
	test.run(t)
}

func TestSame(t *testing.T) {
	test := engineTestCase{
		base:                      testDataGenerator{1234567, 473, nil, nil},
		target:                    testDataGenerator{1234567, 473, nil, nil},
		numberOfOperations:        1,
		expectCoalescedOperations: true,
	}
	test.run(t)
}

func TestSame1Mutation(t *testing.T) {
	test := engineTestCase{
		base:                      testDataGenerator{10240, 473, nil, nil},
		target:                    testDataGenerator{10240, 473, []int{1300}, nil},
		blockSize:                 1024,
		maxDataOpSize:             1024,
		numberOfOperations:        3,
		numberOfDataOperations:    1,
		expectCoalescedOperations: true,
	}
	test.run(t)
}

func TestSame2Mutations(t *testing.T) {
	test := engineTestCase{
		base:                   testDataGenerator{10220, 473, nil, nil},
		target:                 testDataGenerator{10220, 473, []int{2073, 7000}, nil},
		blockSize:              2048,
		maxDataOpSize:          2048,
		numberOfOperations:     5,
		numberOfDataOperations: 2,
	}
	test.run(t)
}

func TestTruncateOnBlockBoundary(t *testing.T) {
	test := engineTestCase{
		base:                      testDataGenerator{999, 212, nil, nil},
		target:                    testDataGenerator{666, 212, nil, nil},
		blockSize:                 333,
		numberOfOperations:        1,
		expectCoalescedOperations: true,
	}
	test.run(t)
}

func TestTruncateOffBlockBoundary(t *testing.T) {
	test := engineTestCase{
		base:                      testDataGenerator{888, 912, nil, nil},
		target:                    testDataGenerator{790, 912, nil, nil},
		blockSize:                 111,
		maxDataOpSize:             1024,
		numberOfOperations:        2,
		numberOfDataOperations:    1,
		expectCoalescedOperations: true,
	}
	test.run(t)
}

func TestPrepend(t *testing.T) {
	test := engineTestCase{
		base:                      testDataGenerator{9880, 11, nil, nil},
		target:                    testDataGenerator{9880, 11, nil, []byte{1, 2, 3}},
		blockSize:                 1234,
		maxDataOpSize:             5,
		numberOfOperations:        2,
		numberOfDataOperations:    1,
		expectCoalescedOperations: true,
	}
	test.run(t)
}

func TestAppend(t *testing.T) {
	test := engineTestCase{
		base:                      testDataGenerator{45271, 473, nil, nil},
		target:                    testDataGenerator{45271 + 876, 473, nil, nil},
		blockSize:                 6453,
		maxDataOpSize:             1024,
		numberOfOperations:        2,
		numberOfDataOperations:    1,
		expectCoalescedOperations: true,
	}
	test.run(t)
}

func TestDifferentDataSameLength(t *testing.T) {
	test := engineTestCase{
		base:                   testDataGenerator{10473, 473, nil, nil},
		target:                 testDataGenerator{10473, 182, nil, nil},
		maxDataOpSize:          1024,
		numberOfOperations:     11,
		numberOfDataOperations: 11,
	}
	test.run(t)
}

func TestDifferentDataDifferentLength(t *testing.T) {
	test := engineTestCase{
		base:                   testDataGenerator{678345, 473, nil, nil},
		target:                 testDataGenerator{473711, 182, nil, nil},
		maxDataOpSize:          12304,
		numberOfOperations:     39,
		numberOfDataOperations: 39,
	}
	test.run(t)
}
