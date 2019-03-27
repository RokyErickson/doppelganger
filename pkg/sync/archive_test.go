package sync

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestArchiveEmptyDifferentEmptyDirectory(t *testing.T) {
	emptyArchive := &Archive{}
	emptyArchiveBytes, err := proto.Marshal(emptyArchive)
	if err != nil {
		t.Fatal("unable to marshal empty archive:", err)
	}
	if len(emptyArchiveBytes) > 0 {
		t.Error("empty archive serialized to non-empty bytes")
	}

	archiveEmptyDirectory := &Archive{Root: &Entry{Kind: EntryKind_Directory}}
	archiveEmptyDirectoryBytes, err := proto.Marshal(archiveEmptyDirectory)
	if err != nil {
		t.Fatal("unable to marshal archive with empty directory:", err)
	}

	if bytes.Equal(emptyArchiveBytes, archiveEmptyDirectoryBytes) {
		t.Error("empty archive and archive with empty directory serialize the same")
	}
}

func TestArchiveConsistentSerialization(t *testing.T) {
	firstEntry := testDirectory1Entry
	secondEntry := firstEntry.Copy()

	firstBuffer := proto.NewBuffer(nil)
	firstBuffer.SetDeterministic(true)
	if err := firstBuffer.Marshal(&Archive{Root: firstEntry}); err != nil {
		t.Fatal("unable to marshal first entry:", err)
	}

	secondBuffer := proto.NewBuffer(nil)
	secondBuffer.SetDeterministic(true)
	if err := secondBuffer.Marshal(&Archive{Root: secondEntry}); err != nil {
		t.Fatal("unable to marshal second entry:", err)
	}

	if !bytes.Equal(firstBuffer.Bytes(), secondBuffer.Bytes()) {
		t.Error("marshalling is not consistent")
	}
}

func TestArchiveNilInvalid(t *testing.T) {
	var archive *Archive

	if archive.EnsureValid() == nil {
		t.Error("nil archive considered valid")
	}
}

func TestArchiveInvalidRootInvalid(t *testing.T) {
	archive := &Archive{
		Root: &Entry{
			Kind:   EntryKind_Directory,
			Digest: []byte{0, 1, 2, 3},
		},
	}

	if archive.EnsureValid() == nil {
		t.Error("archive with invalid root considered valid")
	}
}

func TestArchiveNilRootValid(t *testing.T) {
	archive := &Archive{}

	if err := archive.EnsureValid(); err != nil {
		t.Error("archive with nil root considered invalid:", err)
	}
}

func TestArchiveNonNilRootValid(t *testing.T) {
	archive := &Archive{
		Root: &Entry{
			Kind: EntryKind_Directory,
		},
	}

	if err := archive.EnsureValid(); err != nil {
		t.Error("archive with nil root considered invalid:", err)
	}
}
