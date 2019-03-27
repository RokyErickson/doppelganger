package local

import (
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/session"
)

const (
	numberOfByteValues = 1 << 8
)

type stagingSink struct {
	stager *stager

	path string

	storage *os.File

	digester hash.Hash

	maximumSize uint64

	currentSize uint64
}

func (s *stagingSink) Write(data []byte) (int, error) {

	if s.maximumSize != 0 && (s.maximumSize-s.currentSize) < uint64(len(data)) {
		return 0, errors.New("maximum file size reached")
	}

	n, err := s.storage.Write(data)

	s.digester.Write(data[:n])

	s.currentSize += uint64(n)

	return n, err
}

func (s *stagingSink) Close() error {

	if err := s.storage.Close(); err != nil {
		return errors.Wrap(err, "unable to close underlying storage")
	}

	digest := s.digester.Sum(nil)

	destination, prefix, err := pathForStaging(s.stager.root, s.path, digest)
	if err != nil {
		os.Remove(s.storage.Name())
		return errors.Wrap(err, "unable to compute staging destination")
	}

	if err = s.stager.ensurePrefixExists(prefix); err != nil {
		os.Remove(s.storage.Name())
		return errors.Wrap(err, "unable to create prefix directory")
	}

	if err = os.Rename(s.storage.Name(), destination); err != nil {
		os.Remove(s.storage.Name())
		return errors.Wrap(err, "unable to relocate file")
	}

	return nil
}

type stager struct {
	version         session.Version
	root            string
	maximumFileSize uint64
	rootCreated     bool
	prefixCreated   map[string]bool
}

func newStager(version session.Version, root string, maximumFileSize uint64) *stager {
	return &stager{
		version:         version,
		root:            root,
		maximumFileSize: maximumFileSize,
		prefixCreated:   make(map[string]bool, numberOfByteValues),
	}
}

func (s *stager) ensurePrefixExists(prefix string) error {
	if s.prefixCreated[prefix] {
		return nil
	}

	if err := os.MkdirAll(filepath.Join(s.root, prefix), 0700); err != nil {
		return err
	}
	s.rootCreated = true
	s.prefixCreated[prefix] = true

	return nil
}

func (s *stager) wipe() error {

	s.prefixCreated = make(map[string]bool, numberOfByteValues)

	s.rootCreated = false

	if err := os.RemoveAll(s.root); err != nil {
		errors.Wrap(err, "unable to remove staging directory")
	}

	return nil
}

func (s *stager) Sink(path string) (io.WriteCloser, error) {

	if !s.rootCreated {
		if err := os.MkdirAll(s.root, 0700); err != nil {
			return nil, errors.Wrap(err, "unable to create staging root")
		}
		s.rootCreated = true
	}

	storage, err := ioutil.TempFile(s.root, "staging")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create temporary storage file")
	}

	return &stagingSink{
		stager:      s,
		path:        path,
		storage:     storage,
		digester:    s.version.Hasher(),
		maximumSize: s.maximumFileSize,
	}, nil
}

func (s *stager) Provide(path string, digest []byte) (string, error) {

	expectedLocation, _, err := pathForStaging(s.root, path, digest)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute staging path")
	}

	if _, err := os.Lstat(expectedLocation); err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("file does not exist at expected location")
		}
		return "", errors.Wrap(err, "unable to query staged file metadata")
	}

	return expectedLocation, nil
}
