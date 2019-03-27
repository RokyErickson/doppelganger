package rsync

import (
	"bytes"
	"context"
	"io"

	"github.com/pkg/errors"

	fs "github.com/RokyErickson/doppelganger/pkg/filesystem"
)

func (s *ReceiverStatus) EnsureValid() error {
	if s == nil {
		return nil
	}

	if s.Received > s.Total {
		return errors.New("receiver status indicates too many files received")
	}
	return nil
}

type Receiver interface {
	Receive(*Transmission) error
	finalize() error
}

type Sinker interface {
	Sink(path string) (io.WriteCloser, error)
}

type readSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type emptyReadSeekCloser struct {
	*bytes.Reader
}

func newEmptyReadSeekCloser() readSeekCloser {
	return &emptyReadSeekCloser{bytes.NewReader(nil)}
}

func (e *emptyReadSeekCloser) Close() error {
	return nil
}

type receiver struct {
	root       string
	paths      []string
	signatures []*Signature
	opener     *fs.Opener
	sinker     Sinker
	engine     *Engine
	received   uint64
	total      uint64
	finalized  bool
	burning    bool
	base       readSeekCloser
	target     io.WriteCloser
}

func NewReceiver(root string, paths []string, signatures []*Signature, sinker Sinker) (Receiver, error) {

	if len(paths) != len(signatures) {
		return nil, errors.New("number of paths does not match number of signatures")
	}

	return &receiver{
		root:       root,
		paths:      paths,
		signatures: signatures,
		opener:     fs.NewOpener(root),
		sinker:     sinker,
		engine:     NewEngine(),
		total:      uint64(len(paths)),
	}, nil
}

func (r *receiver) Receive(transmission *Transmission) error {
	if r.finalized {
		panic("receive called on finalized receiver")
	}

	if r.received == r.total {
		return errors.New("unexpected file transmission")
	}

	skip := r.burning

	if transmission.Done {

		if r.base != nil {
			r.base.Close()
			r.base = nil
			r.target.Close()
			r.target = nil
		} else if !r.burning {
			if target, _ := r.sinker.Sink(r.paths[r.received]); target != nil {
				target.Close()
			}
		}

		r.received++

		r.burning = false

		skip = true
	}

	if skip {
		return nil
	}

	signature := r.signatures[r.received]

	if r.base == nil {
		path := r.paths[r.received]

		if signature.isEmpty() {
			r.base = newEmptyReadSeekCloser()
		} else if base, err := r.opener.Open(path); err != nil {
			r.burning = true
			return nil
		} else {
			r.base = base
		}

		if target, err := r.sinker.Sink(path); err != nil {
			r.base.Close()
			r.base = nil
			r.burning = true
			return nil
		} else {
			r.target = target
		}
	}

	if err := r.engine.Patch(r.target, r.base, signature, transmission.Operation); err != nil {
		r.base.Close()
		r.base = nil
		r.target.Close()
		r.target = nil
		r.burning = true
		return nil
	}

	return nil
}

func (r *receiver) finalize() error {
	if r.finalized {
		return errors.New("receiver finalized multiple times")
	}

	if r.base != nil {
		r.base.Close()
		r.base = nil
		r.target.Close()
		r.target = nil
	}

	r.opener.Close()

	r.finalized = true

	return nil
}

type Monitor func(*ReceiverStatus) error

type monitoringReceiver struct {
	receiver  Receiver
	paths     []string
	received  uint64
	total     uint64
	beginning bool
	monitor   Monitor
}

func NewMonitoringReceiver(receiver Receiver, paths []string, monitor Monitor) Receiver {
	return &monitoringReceiver{
		receiver:  receiver,
		paths:     paths,
		total:     uint64(len(paths)),
		beginning: true,
		monitor:   monitor,
	}
}

func (r *monitoringReceiver) Receive(transmission *Transmission) error {
	if err := r.receiver.Receive(transmission); err != nil {
		return err
	}

	if r.received == r.total {
		return errors.New("unexpected file transmission")
	}

	sendStatusUpdate := false

	if r.beginning {
		r.beginning = false
		sendStatusUpdate = true
	}

	if transmission.Done {
		r.received++
		sendStatusUpdate = true
	}

	if sendStatusUpdate {
		var path string
		if r.received < r.total {
			path = r.paths[r.received]
		}

		status := &ReceiverStatus{
			Path:     path,
			Received: r.received,
			Total:    r.total,
		}
		if err := r.monitor(status); err != nil {
			return errors.Wrap(err, "unable to send receiving status")
		}
	}

	return nil
}

func (r *monitoringReceiver) finalize() error {
	r.monitor(nil)
	return r.receiver.finalize()
}

type preemptableReceiver struct {
	receiver Receiver
	run      context.Context
}

func NewPreemptableReceiver(receiver Receiver, run context.Context) Receiver {
	return &preemptableReceiver{
		receiver: receiver,
		run:      run,
	}
}

func (r *preemptableReceiver) Receive(transmission *Transmission) error {

	select {
	case <-r.run.Done():
		return errors.New("reception cancelled")
	default:
	}

	return r.receiver.Receive(transmission)
}

func (r *preemptableReceiver) finalize() error {
	return r.receiver.finalize()
}

type Encoder interface {
	Encode(*Transmission) error
	Finalize() error
}

type encodingReceiver struct {
	encoder   Encoder
	finalized bool
}

func NewEncodingReceiver(encoder Encoder) Receiver {
	return &encodingReceiver{
		encoder: encoder,
	}
}

func (r *encodingReceiver) Receive(transmission *Transmission) error {
	return errors.Wrap(r.encoder.Encode(transmission), "unable to encode transmission")
}

func (r *encodingReceiver) finalize() error {
	if r.finalized {
		return errors.New("receiver finalized multiple times")
	}

	r.finalized = true

	if err := r.encoder.Finalize(); err != nil {
		return errors.Wrap(err, "unable to finalize encoder")
	}

	return nil
}

type Decoder interface {
	Decode(*Transmission) error
	Finalize() error
}

func DecodeToReceiver(decoder Decoder, count uint64, receiver Receiver) error {
	transmission := &Transmission{}

	for count > 0 {
		for {

			transmission.resetToZeroMaintainingCapacity()
			if err := decoder.Decode(transmission); err != nil {
				decoder.Finalize()
				receiver.finalize()
				return errors.Wrap(err, "unable to decode transmission")
			}

			if err := transmission.EnsureValid(); err != nil {
				decoder.Finalize()
				receiver.finalize()
				return errors.Wrap(err, "invalid transmission received")
			}

			if err := receiver.Receive(transmission); err != nil {
				decoder.Finalize()
				receiver.finalize()
				return errors.Wrap(err, "unable to forward message to receiver")
			}

			if transmission.Done {
				break
			}
		}

		count--
	}

	if err := decoder.Finalize(); err != nil {
		receiver.finalize()
		return errors.Wrap(err, "unable to finalize decoder")
	}

	if err := receiver.finalize(); err != nil {
		return errors.Wrap(err, "unable to finalize receiver")
	}

	return nil
}
