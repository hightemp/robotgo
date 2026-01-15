// Copyright (c) 2016-2025 AtomAI, All rights reserved.
//
// See the COPYRIGHT file at the top-level directory of this distribution and at
// https://github.com/go-vgo/robotgo/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0>
//
// This file may not be copied, modified, or distributed
// except according to those terms.

package robotgo

import "testing"

func TestWindowMoveResizeBounds(t *testing.T) {
	if !IsValid() {
		t.Skip("window not valid or accessibility disabled")
	}

	pid := GetPid()
	if pid == 0 {
		t.Skip("no active window pid")
	}

	x, y, w, h := GetBounds(pid)
	if w == 0 || h == 0 {
		t.Skip("invalid window bounds")
	}

	MoveWindow(pid, x, y)
	ResizeWindow(pid, w, h)
	SetWindowBounds(pid, x, y, w, h)

	nx, ny, nw, nh := GetBounds(pid)
	if nw == 0 || nh == 0 {
		t.Fatalf("invalid bounds after calls: %d,%d,%d,%d", nx, ny, nw, nh)
	}
}
