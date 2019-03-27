package daemon

import (
	"encoding/gob"
	"testing"
)

func TestDialTimeoutNoListener(t *testing.T) {
	if c, err := DialTimeout(RecommendedDialTimeout); err == nil {
		c.Close()
		t.Error("IPC connection succeeded unexpectedly")
	}
}

type testIPCMessage struct {
	Name string

	Age uint
}

func TestIPC(t *testing.T) {

	expected := testIPCMessage{"George", 67}

	listener, err := NewListener()
	if err != nil {
		t.Fatal("unable to create listener:", err)
	}
	defer listener.Close()

	go func() {

		connection, err := DialTimeout(RecommendedDialTimeout)
		if err != nil {
			return
		}
		defer connection.Close()

		encoder := gob.NewEncoder(connection)

		encoder.Encode(expected)
	}()

	connection, err := listener.Accept()
	if err != nil {
		t.Fatal("unable to accept connection:", err)
	}
	defer connection.Close()

	decoder := gob.NewDecoder(connection)

	var received testIPCMessage
	if err := decoder.Decode(&received); err != nil {
		t.Fatal("unable to receive test message:", err)
	} else if received != expected {
		t.Error("received message does not match expected:", received, "!=", expected)
	}
}
