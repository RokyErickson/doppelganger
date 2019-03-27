package filesystem

import (
	"os"
	"syscall"
)

func watchRootParametersEqual(first, second os.FileInfo) bool {

	if first == nil && second == nil {
		return true
	} else if first == nil || second == nil {
		return false
	}

	firstData, firstOk := first.Sys().(*syscall.Win32FileAttributeData)
	secondData, secondOk := second.Sys().(*syscall.Win32FileAttributeData)

	return firstOk && secondOk &&
		firstData.FileAttributes == secondData.FileAttributes &&
		firstData.CreationTime == secondData.CreationTime
}
