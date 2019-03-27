package rsync

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"hash"
	"io"
	"math"

	"github.com/pkg/errors"
)

func (h *BlockHash) EnsureValid() error {

	if h == nil {
		return errors.New("nil block hash")
	}

	if len(h.Strong) == 0 {
		return errors.New("empty strong signature")
	}

	return nil
}

func (s *Signature) EnsureValid() error {

	if s == nil {
		return errors.New("nil signature")
	}

	for _, h := range s.Hashes {
		if err := h.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid block hash")
		}
	}

	if s.BlockSize == 0 {
		if s.LastBlockSize != 0 {
			return errors.New("block size of 0 with non-0 last block size")
		} else if len(s.Hashes) != 0 {
			return errors.New("block size of 0 with non-0 number of hashes")
		}
		return nil
	}

	if s.LastBlockSize == 0 {
		return errors.New("non-0 block size with last block size of 0")
	} else if s.LastBlockSize > s.BlockSize {
		return errors.New("last block size greater than block size")
	}

	if len(s.Hashes) == 0 {
		return errors.New("non-0 block size with no block hashes")
	}

	return nil
}

func (s *Signature) isEmpty() bool {

	return s.BlockSize == 0
}

func (o *Operation) EnsureValid() error {

	if o == nil {
		return errors.New("nil operation")
	}

	if len(o.Data) > 0 {
		if o.Start != 0 {
			return errors.New("data operation with non-0 block start index")
		} else if o.Count != 0 {
			return errors.New("data operation with non-0 block count")
		}
	} else if o.Count == 0 {
		return errors.New("block operation with 0 block count")
	}

	return nil
}

func (o *Operation) Copy() *Operation {

	var data []byte
	if len(o.Data) > 0 {
		data = make([]byte, len(o.Data))
		copy(data, o.Data)
	}

	return &Operation{
		Data:  data,
		Start: o.Start,
		Count: o.Count,
	}
}

func (o *Operation) resetToZeroMaintainingCapacity() {

	o.Data = o.Data[:0]

	o.Start = 0
	o.Count = 0
}

func (o *Operation) isZeroValue() bool {
	return len(o.Data) == 0 && o.Start == 0 && o.Count == 0
}

const (
	minimumOptimalBlockSize         = 1 << 10
	maximumOptimalBlockSize         = 1 << 16
	DefaultBlockSize                = 1 << 13
	DefaultMaximumDataOperationSize = 1 << 14
)

func OptimalBlockSizeForBaseLength(baseLength uint64) uint64 {
	result := uint64(math.Sqrt(24.0 * float64(baseLength)))

	if result < minimumOptimalBlockSize {
		result = minimumOptimalBlockSize
	} else if result > maximumOptimalBlockSize {
		result = maximumOptimalBlockSize
	}

	return result
}

func OptimalBlockSizeForBase(base io.Seeker) (uint64, error) {
	if currentOffset, err := base.Seek(0, io.SeekCurrent); err != nil {
		return 0, errors.Wrap(err, "unable to determine current base offset")
	} else if currentOffset < 0 {
		return 0, errors.Wrap(err, "seek return negative starting location")
	} else if length, err := base.Seek(0, io.SeekEnd); err != nil {
		return 0, errors.Wrap(err, "unable to compute base length")
	} else if length < 0 {
		return 0, errors.New("seek returned negative offset")
	} else if _, err = base.Seek(currentOffset, io.SeekStart); err != nil {
		return 0, errors.Wrap(err, "unable to reset base")
	} else {
		return OptimalBlockSizeForBaseLength(uint64(length)), nil
	}
}

type OperationTransmitter func(*Operation) error

type Engine struct {
	buffer           []byte
	strongHasher     hash.Hash
	strongHashBuffer []byte
	targetReader     *bufio.Reader
	operation        *Operation
}

func NewEngine() *Engine {
	strongHasher := sha1.New()

	return &Engine{
		strongHasher:     strongHasher,
		strongHashBuffer: make([]byte, strongHasher.Size()),
		targetReader:     bufio.NewReader(nil),
		operation:        &Operation{},
	}
}

func (e *Engine) bufferWithSize(size uint64) []byte {
	if uint64(cap(e.buffer)) >= size {
		return e.buffer[:size]
	}

	e.buffer = make([]byte, size)
	return e.buffer
}

const (
	m = 1 << 16
)

func (e *Engine) weakHash(data []byte, blockSize uint64) (uint32, uint32, uint32) {
	var r1, r2 uint32
	for i, b := range data {
		r1 += uint32(b)
		r2 += (uint32(blockSize) - uint32(i)) * uint32(b)
	}
	r1 = r1 % m
	r2 = r2 % m

	result := r1 + m*r2

	return result, r1, r2
}

func (e *Engine) rollWeakHash(r1, r2 uint32, out, in byte, blockSize uint64) (uint32, uint32, uint32) {

	r1 = (r1 - uint32(out) + uint32(in)) % m
	r2 = (r2 - uint32(blockSize)*uint32(out) + r1) % m

	result := r1 + m*r2

	return result, r1, r2
}

func (e *Engine) strongHash(data []byte, allocate bool) []byte {

	e.strongHasher.Reset()

	e.strongHasher.Write(data)

	var output []byte
	if !allocate {
		output = e.strongHashBuffer[:0]
	}

	return e.strongHasher.Sum(output)
}

func (e *Engine) Signature(base io.Reader, blockSize uint64) (*Signature, error) {

	if blockSize == 0 {
		if baseSeeker, ok := base.(io.Seeker); ok {
			if s, err := OptimalBlockSizeForBase(baseSeeker); err == nil {
				blockSize = s
			} else {
				blockSize = DefaultBlockSize
			}
		} else {
			blockSize = DefaultBlockSize
		}
	}

	result := &Signature{
		BlockSize: blockSize,
	}

	buffer := e.bufferWithSize(blockSize)

	eof := false
	for !eof {

		n, err := io.ReadFull(base, buffer)
		if err == io.EOF {
			result.LastBlockSize = blockSize
			break
		} else if err == io.ErrUnexpectedEOF {
			result.LastBlockSize = uint64(n)
			eof = true
		} else if err != nil {
			return nil, errors.Wrap(err, "unable to read data block")
		}

		weak, _, _ := e.weakHash(buffer[:n], blockSize)
		strong := e.strongHash(buffer[:n], true)

		result.Hashes = append(result.Hashes, &BlockHash{
			Weak:   weak,
			Strong: strong,
		})
	}

	if len(result.Hashes) == 0 {
		result.BlockSize = 0
		result.LastBlockSize = 0
	}

	return result, nil
}

func (e *Engine) BytesSignature(base []byte, blockSize uint64) *Signature {
	result, err := e.Signature(bytes.NewReader(base), blockSize)
	if err != nil {
		panic(errors.Wrap(err, "in-memory signature failure"))
	}

	return result
}

type dualModeReader interface {
	io.Reader
	io.ByteReader
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func (e *Engine) transmitData(data []byte, transmit OperationTransmitter) error {
	*e.operation = Operation{
		Data: data,
	}

	return transmit(e.operation)
}

func (e *Engine) transmitBlock(start, count uint64, transmit OperationTransmitter) error {

	*e.operation = Operation{
		Start: start,
		Count: count,
	}

	return transmit(e.operation)
}

func (e *Engine) chunkAndTransmitAll(target io.Reader, maxDataOpSize uint64, transmit OperationTransmitter) error {

	if maxDataOpSize == 0 {
		maxDataOpSize = DefaultMaximumDataOperationSize
	}

	buffer := e.bufferWithSize(maxDataOpSize)

	for {
		if n, err := io.ReadFull(target, buffer); err == io.EOF {
			return nil
		} else if err == io.ErrUnexpectedEOF {
			if err = e.transmitData(buffer[:n], transmit); err != nil {
				return errors.Wrap(err, "unable to transmit data operation")
			}
			return nil
		} else if err != nil {
			return errors.Wrap(err, "unable to read target")
		} else if err = e.transmitData(buffer, transmit); err != nil {
			return errors.Wrap(err, "unable to transmit data operation")
		}
	}
}

func (e *Engine) Deltafy(target io.Reader, base *Signature, maxDataOpSize uint64, transmit OperationTransmitter) error {

	if maxDataOpSize == 0 {
		maxDataOpSize = DefaultMaximumDataOperationSize
	}

	if len(base.Hashes) == 0 {
		return e.chunkAndTransmitAll(target, maxDataOpSize, transmit)
	}

	var coalescedStart, coalescedCount uint64
	sendBlock := func(index uint64) error {
		if coalescedCount > 0 {
			if coalescedStart+coalescedCount == index {
				coalescedCount += 1
				return nil
			} else if err := e.transmitBlock(coalescedStart, coalescedCount, transmit); err != nil {
				return nil
			}
		}
		coalescedStart = index
		coalescedCount = 1
		return nil
	}
	sendData := func(data []byte) error {
		if len(data) > 0 && coalescedCount > 0 {
			if err := e.transmitBlock(coalescedStart, coalescedCount, transmit); err != nil {
				return err
			}
			coalescedStart = 0
			coalescedCount = 0
		}
		for len(data) > 0 {
			sendSize := min(uint64(len(data)), maxDataOpSize)
			if err := e.transmitData(data[:sendSize], transmit); err != nil {
				return err
			}
			data = data[sendSize:]
		}
		return nil
	}

	bufferedTarget, ok := target.(dualModeReader)
	if !ok {
		e.targetReader.Reset(target)
		bufferedTarget = e.targetReader
		defer func() {
			e.targetReader.Reset(nil)
		}()
	}

	hashes := base.Hashes
	haveShortLastBlock := false
	var lastBlockIndex uint64
	var shortLastBlock *BlockHash
	if base.LastBlockSize != base.BlockSize {
		haveShortLastBlock = true
		lastBlockIndex = uint64(len(hashes) - 1)
		shortLastBlock = hashes[lastBlockIndex]
		hashes = hashes[:lastBlockIndex]
	}
	weakToBlockHashes := make(map[uint32][]uint64, len(hashes))
	for i, h := range hashes {
		weakToBlockHashes[h.Weak] = append(weakToBlockHashes[h.Weak], uint64(i))
	}

	buffer := e.bufferWithSize(maxDataOpSize + base.BlockSize)

	var occupancy uint64

	var weak, r1, r2 uint32

	for {
		if occupancy == 0 {
			if n, err := io.ReadFull(bufferedTarget, buffer[:base.BlockSize]); err == io.EOF || err == io.ErrUnexpectedEOF {
				occupancy = uint64(n)
				break
			} else if err != nil {
				return errors.Wrap(err, "unable to perform initial buffer fill")
			} else {
				occupancy = base.BlockSize
				weak, r1, r2 = e.weakHash(buffer[:occupancy], base.BlockSize)
			}
		} else if occupancy < base.BlockSize {
			panic("buffer contains less than a block worth of data")
		} else {
			if b, err := bufferedTarget.ReadByte(); err == io.EOF {
				break
			} else if err != nil {
				return errors.Wrap(err, "unable to read target byte")
			} else {
				weak, r1, r2 = e.rollWeakHash(r1, r2, buffer[occupancy-base.BlockSize], b, base.BlockSize)
				buffer[occupancy] = b
				occupancy += 1
			}
		}

		potentials := weakToBlockHashes[weak]
		match := false
		var matchIndex uint64
		if len(potentials) > 0 {
			strong := e.strongHash(buffer[occupancy-base.BlockSize:occupancy], false)
			for _, p := range potentials {
				if bytes.Equal(base.Hashes[p].Strong, strong) {
					match = true
					matchIndex = p
					break
				}
			}
		}

		if match {
			if err := sendData(buffer[:occupancy-base.BlockSize]); err != nil {
				return errors.Wrap(err, "unable to transmit data preceding match")
			} else if err = sendBlock(matchIndex); err != nil {
				return errors.Wrap(err, "unable to transmit match")
			}
			occupancy = 0
		} else if occupancy == uint64(len(buffer)) {
			if err := sendData(buffer[:occupancy-base.BlockSize]); err != nil {
				return errors.Wrap(err, "unable to transmit data before truncation")
			}
			copy(buffer[:base.BlockSize], buffer[occupancy-base.BlockSize:occupancy])
			occupancy = base.BlockSize
		}
	}

	if haveShortLastBlock && occupancy >= base.LastBlockSize {
		potentialLastBlockMatch := buffer[occupancy-base.LastBlockSize : occupancy]
		if w, _, _ := e.weakHash(potentialLastBlockMatch, base.BlockSize); w == shortLastBlock.Weak {
			if bytes.Equal(e.strongHash(potentialLastBlockMatch, false), shortLastBlock.Strong) {
				if err := sendData(buffer[:occupancy-base.LastBlockSize]); err != nil {
					return errors.Wrap(err, "unable to transmit data")
				} else if err = sendBlock(lastBlockIndex); err != nil {
					return errors.Wrap(err, "unable to transmit operation")
				}
				occupancy = 0
			}
		}
	}

	if err := sendData(buffer[:occupancy]); err != nil {
		return errors.Wrap(err, "unable to send final data operation")
	}

	if coalescedCount > 0 {
		if err := e.transmitBlock(coalescedStart, coalescedCount, transmit); err != nil {
			return errors.Wrap(err, "unable to send final block operation")
		}
	}

	return nil
}

func (e *Engine) DeltafyBytes(target []byte, base *Signature, maxDataOpSize uint64) []*Operation {

	var delta []*Operation

	transmit := func(o *Operation) error {
		delta = append(delta, o.Copy())
		return nil
	}

	reader := bytes.NewReader(target)

	if err := e.Deltafy(reader, base, maxDataOpSize, transmit); err != nil {
		panic(errors.Wrap(err, "in-memory deltafication failure"))
	}

	return delta
}

func (e *Engine) Patch(destination io.Writer, base io.ReadSeeker, signature *Signature, operation *Operation) error {

	if len(operation.Data) > 0 {

		if _, err := destination.Write(operation.Data); err != nil {
			return errors.Wrap(err, "unable to write data")
		}
	} else {

		if _, err := base.Seek(int64(operation.Start)*int64(signature.BlockSize), io.SeekStart); err != nil {
			return errors.Wrap(err, "unable to seek to base location")
		}

		for c := uint64(0); c < operation.Count; c++ {

			copyLength := signature.BlockSize
			if operation.Start+c == uint64(len(signature.Hashes)-1) {
				copyLength = signature.LastBlockSize
			}

			buffer := e.bufferWithSize(copyLength)

			if _, err := io.ReadFull(base, buffer); err != nil {
				return errors.Wrap(err, "unable to read block data")
			} else if _, err = destination.Write(buffer); err != nil {
				return errors.Wrap(err, "unable to write block data")
			}
		}
	}
	return nil
}

func (e *Engine) PatchBytes(base []byte, signature *Signature, delta []*Operation) ([]byte, error) {

	baseReader := bytes.NewReader(base)

	output := bytes.NewBuffer(nil)

	for _, o := range delta {
		if err := e.Patch(output, baseReader, signature, o); err != nil {
			return nil, err
		}
	}

	return output.Bytes(), nil
}
