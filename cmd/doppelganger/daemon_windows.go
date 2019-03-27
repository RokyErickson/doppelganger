package main

import (
	"syscall"

	"github.com/RokyErickson/doppelganger/pkg/process"
)

var daemonProcessAttributes = &syscall.SysProcAttr{
	CreationFlags: process.DETACHED_PROCESS | syscall.CREATE_NEW_PROCESS_GROUP,
}
