package filesystem

import (
	"strings"

	"github.com/pkg/errors"
)

var ErrUnsupportedOpenType = errors.New("unsupported open type")

func OpenDirectory(path string, allowSymbolicLinkLeaf bool) (*Directory, *Metadata, error) {
	if d, metadata, err := Open(path, allowSymbolicLinkLeaf); err != nil {
		return nil, nil, err
	} else if (metadata.Mode & ModeTypeMask) != ModeTypeDirectory {
		d.Close()
		return nil, nil, errors.New("path is not a directory")
	} else if directory, ok := d.(*Directory); !ok {
		d.Close()
		panic("invalid directory object returned from open operation")
	} else {
		return directory, metadata, nil
	}
}

func OpenFile(path string, allowSymbolicLinkLeaf bool) (ReadableFile, *Metadata, error) {
	if f, metadata, err := Open(path, allowSymbolicLinkLeaf); err != nil {
		return nil, nil, err
	} else if (metadata.Mode & ModeTypeMask) != ModeTypeFile {
		f.Close()
		return nil, nil, errors.New("path is not a file")
	} else if file, ok := f.(ReadableFile); !ok {
		f.Close()
		panic("invalid file object returned from open operation")
	} else {
		return file, metadata, nil
	}
}

type Opener struct {
	root string

	rootDirectory *Directory

	openParentNames []string

	openParentDirectories []*Directory
}

func NewOpener(root string) *Opener {
	return &Opener{root: root}
}

func (o *Opener) Open(path string) (ReadableFile, error) {

	if path == "" {

		if o.rootDirectory != nil {
			return nil, errors.New("root already opened as directory")
		}

		if file, _, err := OpenFile(o.root, false); err != nil {
			return nil, errors.Wrap(err, "unable to open root file")
		} else {
			return file, nil
		}
	}

	components := strings.Split(path, "/")
	parentComponents := components[:len(components)-1]
	leafName := components[len(components)-1]

	if o.rootDirectory == nil {
		if directory, _, err := OpenDirectory(o.root, false); err != nil {
			return nil, errors.Wrap(err, "unable to open root directory")
		} else {
			o.rootDirectory = directory
		}
	}

	parent := o.rootDirectory

	for c, component := range parentComponents {

		if c < len(o.openParentNames) {
			if o.openParentNames[c] == component {
				parent = o.openParentDirectories[c]
				continue
			} else {
				for i := c; i < len(o.openParentNames); i++ {

					if err := o.openParentDirectories[i].Close(); err != nil {
						return nil, errors.Wrap(err, "unable to close previous parent directory")
					}

					o.openParentNames[i] = ""
					o.openParentDirectories[i] = nil
				}
				o.openParentNames = o.openParentNames[:c]
				o.openParentDirectories = o.openParentDirectories[:c]
			}
		}

		if directory, err := parent.OpenDirectory(component); err != nil {
			return nil, errors.Wrap(err, "unable to open parent directory")
		} else {
			parent = directory
			o.openParentNames = append(o.openParentNames, component)
			o.openParentDirectories = append(o.openParentDirectories, directory)
		}
	}

	return parent.OpenFile(leafName)
}

func (o *Opener) Close() error {

	var firstErr error

	if o.rootDirectory != nil {
		firstErr = o.rootDirectory.Close()
	}

	for _, directory := range o.openParentDirectories {
		if directory == nil {
			continue
		} else if err := directory.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
