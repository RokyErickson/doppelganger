package rsync

import (
	"github.com/pkg/errors"

	fs "github.com/RokyErickson/doppelganger/pkg/filesystem"
)

func Transmit(root string, paths []string, signatures []*Signature, receiver Receiver) error {

	if len(paths) != len(signatures) {
		receiver.finalize()
		return errors.New("number of paths does not match number of signatures")
	}
	opener := fs.NewOpener(root)
	defer opener.Close()
	engine := NewEngine()
	transmission := &Transmission{}
	for i, p := range paths {
		file, err := opener.Open(p)
		if err != nil {
			*transmission = Transmission{
				Done:  true,
				Error: errors.Wrap(err, "unable to open file").Error(),
			}
			if err = receiver.Receive(transmission); err != nil {
				receiver.finalize()
				return errors.Wrap(err, "unable to send error transmission")
			}
			continue
		}
		var transmitError error
		transmit := func(o *Operation) error {
			*transmission = Transmission{Operation: o}
			transmitError = receiver.Receive(transmission)
			return transmitError
		}
		err = engine.Deltafy(file, signatures[i], 0, transmit)

		file.Close()

		if transmitError != nil {
			receiver.finalize()
			return errors.Wrap(transmitError, "unable to transmit delta")
		}

		*transmission = Transmission{Done: true}
		if err != nil {
			transmission.Error = errors.Wrap(err, "engine error").Error()
		}
		if err = receiver.Receive(transmission); err != nil {
			receiver.finalize()
			return errors.Wrap(err, "unable to send done message")
		}
	}

	if err := receiver.finalize(); err != nil {
		return errors.Wrap(err, "unable to finalize receiver")
	}

	return nil
}
