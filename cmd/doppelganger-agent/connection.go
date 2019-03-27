package main

import (
	"io"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
)

type stdioAddress struct{}

func (stdioAddress) Network() string {
	return "standard input/output"
}

func (stdioAddress) String() string {
	return "standard input/output"
}

type stdioConnection struct {
	io.Reader
	io.Writer
}

func newStdioConnection() *stdioConnection {
	return &stdioConnection{os.Stdin, os.Stdout}
}

func (c *stdioConnection) Close() error {

	return errors.New("closing standard input/output connection not allowed")
}

func (c *stdioConnection) LocalAddr() net.Addr {
	return stdioAddress{}
}

func (c *stdioConnection) RemoteAddr() net.Addr {
	return stdioAddress{}
}

func (c *stdioConnection) SetDeadline(_ time.Time) error {
	return errors.New("deadlines not supported by standard input/output connections")
}

func (c *stdioConnection) SetReadDeadline(_ time.Time) error {
	return errors.New("read deadlines not supported by standard input/output connections")
}

func (c *stdioConnection) SetWriteDeadline(_ time.Time) error {
	return errors.New("write deadlines not supported by standard input/output connections")
}
