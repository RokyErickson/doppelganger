// Subset of https://github.com/rjeczalik/notify extracted and modified to
// expose watcher functionality directly. Originally extracted from the
// following revision:
// https://github.com/rjeczalik/notify/tree/52ae50d8490436622a8941bd70c3dbe0acdd4bbf
//
// The original code license:
//
// The MIT License (MIT)
//
// Copyright (c) 2014-2015 The Notify Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// The original license header inside the code itself:
//
// Copyright (c) 2014-2015 The Notify Authors. All rights reserved.
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

// +build linux

package notify

import (
	"fmt"
	"strings"
)

type Event uint32

const (
	Create = osSpecificCreate
	Remove = osSpecificRemove
	Write  = osSpecificWrite
	Rename = osSpecificRename

	All = Create | Remove | Write | Rename
)

const internal = recursive | omit

func (e Event) String() string {
	var s []string
	for _, strmap := range []map[Event]string{estr, osestr} {
		for ev, str := range strmap {
			if e&ev == ev {
				s = append(s, str)
			}
		}
	}
	return strings.Join(s, "|")
}

type EventInfo interface {
	Event() Event
	Path() string
	Sys() interface{}
}

type isDirer interface {
	isDir() (bool, error)
}

var _ fmt.Stringer = (*event)(nil)
var _ isDirer = (*event)(nil)

func (e *event) String() string {
	return e.Event().String() + `: "` + e.Path() + `"`
}

var estr = map[Event]string{
	Create:    "notify.Create",
	Remove:    "notify.Remove",
	Write:     "notify.Write",
	Rename:    "notify.Rename",
	recursive: "recursive",
	omit:      "omit",
}
