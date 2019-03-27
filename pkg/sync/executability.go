package sync

import (
	"bytes"
)

func propagateExecutabilityRecursive(ancestor, source, target *Entry) {

	if (ancestor == nil && source == nil) || target == nil {
		return
	}

	if target.Kind == EntryKind_Directory {

		ancestorContents := ancestor.GetContents()
		sourceContents := source.GetContents()
		targetContents := target.GetContents()

		if len(sourceContents) == 0 && len(ancestorContents) == 0 {
			return
		}

		for name := range targetContents {
			propagateExecutabilityRecursive(ancestorContents[name], sourceContents[name], targetContents[name])
		}
	} else if target.Kind == EntryKind_File {

		propagateFromSource := source != nil && source.Kind == EntryKind_File &&
			bytes.Equal(source.Digest, target.Digest)
		if propagateFromSource {
			target.Executable = source.Executable
			return
		}

		propagateFromAncestor := ancestor != nil && ancestor.Kind == EntryKind_File &&
			bytes.Equal(ancestor.Digest, target.Digest)
		if propagateFromAncestor {
			target.Executable = ancestor.Executable
			return
		}

		propagateFromSource = source != nil && ancestor != nil &&
			source.Kind == EntryKind_File && ancestor.Kind == EntryKind_File &&
			bytes.Equal(source.Digest, ancestor.Digest)
		if propagateFromSource {
			target.Executable = source.Executable
			return
		}

	}
}

func PropagateExecutability(ancestor, source, target *Entry) *Entry {

	result := target.Copy()

	propagateExecutabilityRecursive(ancestor, source, result)

	return result
}
