package encoding

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

func loadAndUnmarshal(path string, unmarshal func([]byte) error) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return errors.Wrap(err, "unable to load file")
	}

	if err := unmarshal(data); err != nil {
		return errors.Wrap(err, "unable to unmarshal data")
	}

	return nil
}

func marshalAndSave(path string, marshal func() ([]byte, error)) error {

	data, err := marshal()
	if err != nil {
		return errors.Wrap(err, "unable to marshal message")
	}

	if err := filesystem.WriteFileAtomic(path, data, 0600); err != nil {
		return errors.Wrap(err, "unable to write message data")
	}

	return nil
}
