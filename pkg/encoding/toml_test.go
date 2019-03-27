package encoding

import (
	"io/ioutil"
	"os"
	"testing"
)

type testMessageTOML struct {
	Section struct {
		Name string
		Age  uint
	}
}

const (
	testMessageTOMLString = "[section]\nname=\"Abraham\"\nage=56\n"

	testMessageTOMLName = "Abraham"

	testMessageTOMLAge = 56
)

func TestLoadAndUnmarshalTOML(t *testing.T) {

	file, err := ioutil.TempFile("", "doppelganger_encoding")
	if err != nil {
		t.Fatal("unable to create temporary file:", err)
	} else if _, err = file.Write([]byte(testMessageTOMLString)); err != nil {
		t.Fatal("unable to write data to temporary file:", err)
	} else if err = file.Close(); err != nil {
		t.Fatal("unable to close temporary file:", err)
	}
	defer os.Remove(file.Name())

	value := &testMessageTOML{}
	if err := LoadAndUnmarshalTOML(file.Name(), value); err != nil {
		t.Fatal("loadAndUnmarshal failed:", err)
	}

	if value.Section.Name != testMessageTOMLName {
		t.Error("test message name mismatch:", value.Section.Name, "!=", testMessageTOMLName)
	}
	if value.Section.Age != testMessageTOMLAge {
		t.Error("test message age mismatch:", value.Section.Age, "!=", testMessageTOMLAge)
	}
}
