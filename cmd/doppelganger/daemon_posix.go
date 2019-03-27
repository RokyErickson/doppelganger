// +build !windows,!plan9

package main

import (
	"syscall"
)

var daemonProcessAttributes = &syscall.SysProcAttr{
	Setsid: true,
}
