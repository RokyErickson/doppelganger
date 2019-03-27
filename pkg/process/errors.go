package process

import (
	"strings"
)

const (
	windowsInvalidCommandFragment  = "is not recognized as an internal or external command"
	windowsCommandNotFoundFragment = "The system cannot find the path specified"
)

func OutputIsWindowsInvalidCommand(output string) bool {
	return strings.Contains(output, windowsInvalidCommandFragment)
}

func OutputIsWindowsCommandNotFound(output string) bool {
	return strings.Contains(output, windowsCommandNotFoundFragment)
}
