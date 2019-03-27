package encoding

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

type testMessageJSON struct {
	Name string

	Age uint
}

const (
	testMessageJSONString = `{"Name":"George","Age":67}`

	testMessageJSONName = "George"

	testMessageJSONAge = 67
)

func TestLoadAndUnmarshalNonExistentPath(t *testing.T) {
	if !os.IsNotExist(loadAndUnmarshal("/this/does/not/exist", nil)) {
		t.Error("expected loadAndUnmarshal to pass through non-existence errors")
	}
}

func TestLoadAndUnmarshalDirectory(t *testing.T) {
	if loadAndUnmarshal(filesystem.HomeDirectory, nil) == nil {
		t.Error("expected loadAndUnmarshal error when loading directory")
	}
}

func TestLoadAndUnmarshalUnmarshalFail(t *testing.T) {

	file, err := ioutil.TempFile("", "doppelganger_encoding")
	if err != nil {
		t.Fatal("unable to create temporary file:", err)
	} else if err = file.Close(); err != nil {
		t.Fatal("unable to close temporary file:", err)
	}
	defer os.Remove(file.Name())

	unmarshal := func(_ []byte) error {
		return errors.New("unmarshal failed")
	}

	if loadAndUnmarshal(file.Name(), unmarshal) == nil {
		t.Error("expected loadAndUnmarshal to return an error")
	}
}

func TestLoadAndUnmarshal(t *testing.T) {

	file, err := ioutil.TempFile("", "doppelganger_encoding")
	if err != nil {
		t.Fatal("unable to create temporary file:", err)
	} else if _, err = file.Write([]byte(testMessageJSONString)); err != nil {
		t.Fatal("unable to write data to temporary file:", err)
	} else if err = file.Close(); err != nil {
		t.Fatal("unable to close temporary file:", err)
	}
	defer os.Remove(file.Name())

	value := &testMessageJSON{}
	unmarshal := func(data []byte) error {
		return json.Unmarshal(data, value)
	}

	if err := loadAndUnmarshal(file.Name(), unmarshal); err != nil {
		t.Fatal("loadAndUnmarshal failed:", err)
	}

	if value.Name != testMessageJSONName {
		t.Error("test message name mismatch:", value.Name, "!=", testMessageJSONName)
	}
	if value.Age != testMessageJSONAge {
		t.Error("test message age mismatch:", value.Age, "!=", testMessageJSONAge)
	}
}

func TestMarshalAndSaveMarshalFail(t *testing.T) {

	file, err := ioutil.TempFile("", "doppelganger_encoding")
	if err != nil {
		t.Fatal("unable to create temporary file:", err)
	} else if err = file.Close(); err != nil {
		t.Fatal("unable to close temporary file:", err)
	}
	defer os.Remove(file.Name())

	marshal := func() ([]byte, error) {
		return nil, errors.New("marshal failed")
	}

	if marshalAndSave(file.Name(), marshal) == nil {
		t.Error("expected marshalAndSave to return an error")
	}
}

func TestMarshalAndSaveInvalidPath(t *testing.T) {

	directory, err := ioutil.TempDir("", "doppelganger_encoding")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(directory)

	marshal := func() ([]byte, error) {
		return []byte{0}, nil
	}

	if marshalAndSave(directory, marshal) == nil {
		t.Error("expected marshalAndSave to return an error")
	}
}

func TestMarshalAndSave(t *testing.T) {

	file, err := ioutil.TempFile("", "doppelganger_encoding")
	if err != nil {
		t.Fatal("unable to create temporary file:", err)
	} else if err = file.Close(); err != nil {
		t.Fatal("unable to close temporary file:", err)
	}
	defer os.Remove(file.Name())

	value := &testMessageJSON{Name: testMessageJSONName, Age: testMessageJSONAge}
	marshal := func() ([]byte, error) {
		return json.Marshal(value)
	}

	if err := marshalAndSave(file.Name(), marshal); err != nil {
		t.Fatal("marshalAndSave failed:", err)
	}

	contents, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Fatal("unable to read saved contents:", err)
	} else if string(contents) != testMessageJSONString {
		t.Error("marshaled contents do not match expected:", string(contents), "!=", testMessageJSONString)
	}
}
