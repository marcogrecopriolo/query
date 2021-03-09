//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

// +build linux

package vliner

import (
	"syscall"
)

const (
	_TCGETS = syscall.TCGETS
	_TCSETS = syscall.TCSETS
)

type Termios struct {
	syscall.Termios
}
