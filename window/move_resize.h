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

#ifndef MOVE_RESIZE_H
#define MOVE_RESIZE_H

// MoveWindow - move window
void move_window(uintptr pid, int32_t x, int32_t y, int8_t isPid){
	if (!is_valid()) { return; }

	#if defined(IS_MACOSX)
		AXUIElementRef AxID = AXUIElementCreateApplication(pid);
		AXUIElementRef AxWin = NULL;
		if (AxID == NULL) { return; }

		// Get focused window, fall back to main window.
		if (AXUIElementCopyAttributeValue(AxID, kAXFocusedWindowAttribute,
			(CFTypeRef *)&AxWin) != kAXErrorSuccess || AxWin == NULL) {
			AXUIElementCopyAttributeValue(AxID, kAXMainWindowAttribute,
				(CFTypeRef *)&AxWin);
		}

		if (AxWin != NULL) {
			CGPoint newPoint = CGPointMake(x, y);
			AXValueRef position = AXValueCreate(kAXValueCGPointType, &newPoint);
			if (position != NULL) {
				AXUIElementSetAttributeValue(AxWin, kAXPositionAttribute, position);
				CFRelease(position);
			}
			CFRelease(AxWin);
		}
		CFRelease(AxID);
	#elif defined(USE_X11)
		Display *display = XOpenDisplay(NULL);
		if (display == NULL) { return; }
		Window xwin = (Window)pid;

		XMoveWindow(display, xwin, x, y);
		XFlush(display);
		XCloseDisplay(display);
	#elif defined(IS_WINDOWS)
		HWND hwnd = getHwnd(pid, isPid);
		if (hwnd == NULL) { return; }

		SetWindowPos(hwnd, HWND_TOP, x, y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
	#endif
}

// ResizeWindow - resize window
void resize_window(uintptr pid, int32_t width, int32_t height, int8_t isPid){
	if (!is_valid()) { return; }

	#if defined(IS_MACOSX)
		AXUIElementRef AxID = AXUIElementCreateApplication(pid);
		AXUIElementRef AxWin = NULL;
		if (AxID == NULL) { return; }

		if (AXUIElementCopyAttributeValue(AxID, kAXFocusedWindowAttribute,
			(CFTypeRef *)&AxWin) != kAXErrorSuccess || AxWin == NULL) {
			AXUIElementCopyAttributeValue(AxID, kAXMainWindowAttribute,
				(CFTypeRef *)&AxWin);
		}

		if (AxWin != NULL) {
			CGSize newSize = CGSizeMake(width, height);
			AXValueRef size = AXValueCreate(kAXValueCGSizeType, &newSize);
			if (size != NULL) {
				AXUIElementSetAttributeValue(AxWin, kAXSizeAttribute, size);
				CFRelease(size);
			}
			CFRelease(AxWin);
		}
		CFRelease(AxID);
	#elif defined(USE_X11)
		Display *display = XOpenDisplay(NULL);
		if (display == NULL) { return; }
		Window xwin = (Window)pid;

		XResizeWindow(display, xwin, width, height);
		XFlush(display);
		XCloseDisplay(display);
	#elif defined(IS_WINDOWS)
		HWND hwnd = getHwnd(pid, isPid);
		if (hwnd == NULL) { return; }

		SetWindowPos(hwnd, HWND_TOP, 0, 0, width, height, SWP_NOMOVE | SWP_NOZORDER);
	#endif
}

// SetBounds - move and resize window
void set_window_bounds(uintptr pid, int32_t x, int32_t y,
	int32_t width, int32_t height, int8_t isPid){
	if (!is_valid()) { return; }

	#if defined(IS_MACOSX)
		AXUIElementRef AxID = AXUIElementCreateApplication(pid);
		AXUIElementRef AxWin = NULL;
		if (AxID == NULL) { return; }

		if (AXUIElementCopyAttributeValue(AxID, kAXFocusedWindowAttribute,
			(CFTypeRef *)&AxWin) != kAXErrorSuccess || AxWin == NULL) {
			AXUIElementCopyAttributeValue(AxID, kAXMainWindowAttribute,
				(CFTypeRef *)&AxWin);
		}

		if (AxWin != NULL) {
			CGPoint newPoint = CGPointMake(x, y);
			CGSize newSize = CGSizeMake(width, height);
			AXValueRef position = AXValueCreate(kAXValueCGPointType, &newPoint);
			AXValueRef size = AXValueCreate(kAXValueCGSizeType, &newSize);

			if (position != NULL) {
				AXUIElementSetAttributeValue(AxWin, kAXPositionAttribute, position);
				CFRelease(position);
			}
			if (size != NULL) {
				AXUIElementSetAttributeValue(AxWin, kAXSizeAttribute, size);
				CFRelease(size);
			}
			CFRelease(AxWin);
		}
		CFRelease(AxID);
	#elif defined(USE_X11)
		Display *display = XOpenDisplay(NULL);
		if (display == NULL) { return; }
		Window xwin = (Window)pid;

		XMoveResizeWindow(display, xwin, x, y, width, height);
		XFlush(display);
		XCloseDisplay(display);
	#elif defined(IS_WINDOWS)
		HWND hwnd = getHwnd(pid, isPid);
		if (hwnd == NULL) { return; }

		SetWindowPos(hwnd, HWND_TOP, x, y, width, height, SWP_NOZORDER);
	#endif
}

#endif
