package process

import (
	"github.com/pkg/errors"
	"github.com/polydawn/gosh"
	"io"
	"net"
	"sync"
	"time"
)

const smallBufferSize = 64

type Buffer1 struct {
	buf      []byte
	off      int
	lastRead readOp
}

const maxInt = int(^uint(0) >> 1)

type readOp int8

const (
	opRead    readOp = -1
	opInvalid readOp = 0
)

var ErrTooLarge = errors.New("bytes.Buffer: too large")

type address struct{}

func (_ address) Network() string {
	return "standard input/output"
}

func (_ address) String() string {
	return "standard input/output"
}

type Connection struct {
	process gosh.Proc

	killDelayLock sync.Mutex

	killDelay time.Duration
	Buffer1   *Buffer1
}

func NewConnection(process gosh.Command, killDelay time.Duration) (*Connection, error) {

	if killDelay < time.Duration(0) {
		panic("negative kill delay specified")
	}

	return &Connection{
		process:   process.Run(),
		killDelay: killDelay,
		Buffer1: &Buffer1{
			buf:      make([]byte, 0),
			off:      0,
			lastRead: opInvalid,
		},
	}, nil
}

func (c *Connection) Read(buffer []byte) (int, error) {
	return c.Buffer1.read1(buffer)
}

func (c *Connection) Write(buffer []byte) (int, error) {
	return c.Buffer1.write1(buffer)
}

func (c *Connection) SetKillDelay(killDelay time.Duration) {

	if killDelay < time.Duration(0) {
		panic("negative kill delay specified")
	}

	c.killDelayLock.Lock()
	defer c.killDelayLock.Unlock()

	c.killDelay = killDelay
}

func (c *Connection) Close() error {

	if c.process == nil {
		return errors.New("process not started")
	}

	c.killDelayLock.Lock()
	killDelay := c.killDelay
	c.killDelayLock.Unlock()

	c.process.WaitSoon(killDelay)
	c.process.Kill()

	return nil
}

func (c Connection) LocalAddr() net.Addr {
	return address{}
}

func (c Connection) RemoteAddr() net.Addr {
	return address{}
}

func (c Connection) SetDeadline(_ time.Time) error {
	return errors.New("deadlines not supported by process connections")
}

func (c Connection) SetReadDeadline(_ time.Time) error {
	return errors.New("read deadlines not supported by process connections")
}

func (c Connection) SetWriteDeadline(_ time.Time) error {
	return errors.New("write deadlines not supported by process connections")
}

func (b *Buffer1) empty1() bool { return len(b.buf) <= b.off }

func (b *Buffer1) reset1() {
	b.buf = b.buf[:0]
	b.off = 0
	b.lastRead = opInvalid
}

func (b *Buffer1) read1(p []byte) (n int, err error) {
	b.lastRead = opInvalid
	if b.empty1() {
		b.reset1()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	if n > 0 {
		b.lastRead = opRead
	}
	return n, nil
}

func (b *Buffer1) write1(p []byte) (n int, err error) {
	b.lastRead = opInvalid
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

func (b *Buffer1) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

func (b *Buffer1) grow(n int) int {
	m := b.len1()
	if m == 0 && b.off != 0 {
		b.reset1()
	}
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		buf := makeSlice(2*c + n)
		copy(buf, b.buf[b.off:])
		b.buf = buf
	}
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

func (b *Buffer1) len1() int { return len(b.buf) - b.off }

func makeSlice(n int) []byte {
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}
