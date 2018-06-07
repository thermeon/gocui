// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"io"
)

// A View is a window. It maintains its own internal buffer and cursor
// position.
type Viewer interface {
	Size() (x, y int)
	Name() string
	SetCursor(x, y int) error
	Cursor() (x, y int)
	SetOrigin(x, y int) error
	Origin() (x, y int)
	io.Writer
	io.Reader
	Rewind()
	Clear()
	Buffer() string
	ViewBuffer() string
	Line(y int) (string, error)
	Word(x, y int) (string, error)
	Invalidate()
	HasFrame() bool
	SetFrame(f bool)
	GetTitle() string
	SetTitle(t string)
	ClearRunes()
	Draw() error
	GetEditor() Editor
	SetEditor(e Editor)
	IsEditable() bool
	SetEditable(e bool)
	GetBounds() (x0, y0, x1, y1 int)
	SetBounds(x0, y0, x1, y1 int)
	SetBgFgColor(bg Attribute, fg Attribute)
	SetSelBgFgColor(bg Attribute, fg Attribute)
	SetHighlight(h bool)
	GetHighlight() bool
	EditWrite(ch rune)
	EditDelete(back bool)
	EditNewLine()
	MoveCursor(dx, dy int, writeMode bool)
	GetOverwrite() bool
	SetOverwrite(o bool)
	SetMask(m rune)
	GetMask() rune
	GetWrap() bool
	SetWrap(b bool)
	GetAutoscroll() bool
	SetAutoscroll(b bool)
}
